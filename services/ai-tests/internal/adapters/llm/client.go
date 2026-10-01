package llm

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/ports"
)

const (
	promptVersion   = "adaptive-v2"
	maxResponseSize = 4 << 20
	requestTimeout  = 90 * time.Second
)

const roundSystemPrompt = "You generate adaptive educational tests.\n" +
	"Your response is consumed by a Go backend.\n" +
	"Follow the requested JSON contract exactly.\n" +
	"Return JSON only. Do not add Markdown, comments, explanations outside JSON, or extra fields.\n" +
	"All numeric fields must be JSON numbers, never strings."

const feedbackSystemPrompt = "You generate concise educational feedback.\n" +
	"Your response is consumed by a Go backend.\n" +
	"Follow the requested JSON contract exactly.\n" +
	"Return JSON only. Do not add Markdown, comments, explanations outside JSON, or extra fields.\n" +
	"All numeric fields must be JSON numbers, never strings."

type Client struct {
	logger  *slog.Logger
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

func NewClient(logger *slog.Logger, baseURL, apiKey, model string) *Client {
	return &Client{
		logger:  logger,
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		http:    &http.Client{Timeout: requestTimeout},
	}
}

func (c *Client) GenerateRound(ctx context.Context, in ports.GenerateRoundRequest) (ports.GeneratedRound, ports.GenerationMetadata, error) {
	prompt := buildRoundPrompt(in)

	content, inputTokens, outputTokens, err := c.complete(ctx, roundSystemPrompt, prompt)
	if err != nil {
		return ports.GeneratedRound{}, ports.GenerationMetadata{}, err
	}

	var generated ports.GeneratedRound
	if err := decodeJSON(content, &generated); err != nil {
		return ports.GeneratedRound{}, ports.GenerationMetadata{}, fmt.Errorf("decode generated round: %w", err)
	}
	if err := validateGeneratedRound(generated, in); err != nil {
		return ports.GeneratedRound{}, ports.GenerationMetadata{}, fmt.Errorf("validate generated round: %w", err)
	}

	sum := sha256.Sum256([]byte(prompt))

	return generated, ports.GenerationMetadata{
		Model:         c.model,
		PromptVersion: promptVersion,
		PromptHash:    hex.EncodeToString(sum[:]),
		InputTokens:   inputTokens,
		OutputTokens:  outputTokens,
	}, nil
}

func (c *Client) GenerateFeedback(ctx context.Context, in ports.GenerateFeedbackRequest) (ports.GeneratedFeedback, error) {
	prompt := buildFeedbackPrompt(in)

	content, _, _, err := c.complete(ctx, feedbackSystemPrompt, prompt)
	if err != nil {
		return ports.GeneratedFeedback{}, err
	}

	var generated ports.GeneratedFeedback
	if err := decodeJSON(content, &generated); err != nil {
		return ports.GeneratedFeedback{}, fmt.Errorf("decode generated feedback: %w", err)
	}
	if err := validateGeneratedFeedback(generated, in); err != nil {
		return ports.GeneratedFeedback{}, fmt.Errorf("validate generated feedback: %w", err)
	}

	return generated, nil
}

func (c *Client) complete(ctx context.Context, systemPrompt, prompt string) (string, *int, *int, error) {
	start := time.Now()

	apiKey := strings.TrimSpace(c.apiKey)
	if apiKey == "" {
		return "", nil, nil, fmt.Errorf("LLM_API_KEY is empty")
	}

	c.logger.Info("sending llm request", "model", c.model, "auth_configured", true)

	payload := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"thinking": map[string]string{
			"type": "disabled",
		},
		"max_tokens": 4096,
		"temperature": 0.2,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", nil, nil, fmt.Errorf("marshal llm request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", nil, nil, fmt.Errorf("create llm request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		c.logger.Error(
			"llm request failed",
			"model", c.model,
			"duration_ms", time.Since(start).Milliseconds(),
			"error", err,
		)
		return "", nil, nil, fmt.Errorf("llm request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return "", nil, nil, fmt.Errorf("read llm response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		c.logger.Error(
			"llm returned non-success status",
			"model", c.model,
			"status", resp.StatusCode,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return "", nil, nil, fmt.Errorf(
			"llm returned status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(data)),
		)
	}

	content, finishReason, inputTokens, outputTokens, err := parseLLMResponse(data)
	if err != nil {
		return "", nil, nil, err
	}

	if finishReason == "length" {
		return "", nil, nil, fmt.Errorf("llm response truncated: max_tokens reached")
	}

	if strings.TrimSpace(content) == "" {
		return "", nil, nil, fmt.Errorf("llm returned no content")
	}

	c.logger.Info(
		"llm request completed",
		"model", c.model,
		"status", resp.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
		"finish_reason", finishReason,
		"input_tokens", inputTokens,
		"output_tokens", outputTokens,
	)

	return strings.TrimSpace(content), &inputTokens, &outputTokens, nil
}

func parseLLMResponse(data []byte) (string, string, int, int, error) {
	var envelope map[string]json.RawMessage

	if err := json.Unmarshal(data, &envelope); err != nil {
		return "", "", 0, 0, fmt.Errorf("decode llm response: %w", err)
	}

	var choices []map[string]json.RawMessage
	if err := json.Unmarshal(envelope["choices"], &choices); err != nil {
		return "", "", 0, 0, fmt.Errorf("decode llm choices: %w", err)
	}
	if len(choices) == 0 {
		return "", "", 0, 0, fmt.Errorf("llm returned no choices")
	}

	var finishReason string
	if err := json.Unmarshal(choices[0]["finish_reason"], &finishReason); err != nil {
		return "", "", 0, 0, fmt.Errorf("decode llm finish reason: %w", err)
	}

	var message map[string]json.RawMessage
	if err := json.Unmarshal(choices[0]["message"], &message); err != nil {
		return "", "", 0, 0, fmt.Errorf("decode llm message: %w", err)
	}

	var content string
	if err := json.Unmarshal(message["content"], &content); err != nil {
		return "", "", 0, 0, fmt.Errorf("decode llm content: %w", err)
	}

	var usage struct {
		PromptTokens     int
		CompletionTokens int
	}

	if rawUsage, ok := envelope["usage"]; ok {
		var usageRaw map[string]json.RawMessage
		if err := json.Unmarshal(rawUsage, &usageRaw); err != nil {
			return "", "", 0, 0, fmt.Errorf("decode llm usage: %w", err)
		}

		if value, ok := usageRaw["prompt_tokens"]; ok {
			if err := json.Unmarshal(value, &usage.PromptTokens); err != nil {
				return "", "", 0, 0, fmt.Errorf("decode prompt token count: %w", err)
			}
		}
		if value, ok := usageRaw["completion_tokens"]; ok {
			if err := json.Unmarshal(value, &usage.CompletionTokens); err != nil {
				return "", "", 0, 0, fmt.Errorf("decode completion token count: %w", err)
			}
		}
	}

	return content, finishReason, usage.PromptTokens, usage.CompletionTokens, nil
}

func decodeJSON(content string, target any) error {
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(content)))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return fmt.Errorf("trailing data after JSON: %w", err)
	}

	return nil
}

func validateGeneratedRound(round ports.GeneratedRound, in ports.GenerateRoundRequest) error {
	if len(round.Questions) != in.QuestionCount {
		return fmt.Errorf("expected %d questions, got %d", in.QuestionCount, len(round.Questions))
	}

	allowedTopics := make(map[int64]string, len(in.AllowedTopics))
	for _, topic := range in.AllowedTopics {
		allowedTopics[topic.PageID] = topic.Title
	}

	for i, question := range round.Questions {
		if question.Difficulty < 1 || question.Difficulty > 5 {
			return fmt.Errorf("question %d: difficulty must be between 1 and 5, got %d", i+1, question.Difficulty)
		}

		if len(question.Options) != 4 {
			return fmt.Errorf("question %d: expected 4 options, got %d", i+1, len(question.Options))
		}

		seenKeys := make(map[string]struct{}, 4)
		for _, option := range question.Options {
			if option.Key != "A" && option.Key != "B" && option.Key != "C" && option.Key != "D" {
				return fmt.Errorf("question %d: invalid option key %q", i+1, option.Key)
			}
			if _, exists := seenKeys[option.Key]; exists {
				return fmt.Errorf("question %d: duplicate option key %q", i+1, option.Key)
			}
			if strings.TrimSpace(option.Text) == "" {
				return fmt.Errorf("question %d: option %s is empty", i+1, option.Key)
			}
			seenKeys[option.Key] = struct{}{}
		}

		if len(seenKeys) != 4 {
			return fmt.Errorf("question %d: options must contain A, B, C and D", i+1)
		}
		if _, ok := seenKeys[question.CorrectOptionKey]; !ok {
			return fmt.Errorf("question %d: correctOptionKey %q does not match an option", i+1, question.CorrectOptionKey)
		}

		if question.TopicPageID == nil {
			if question.TopicTitle != "" {
				return fmt.Errorf("question %d: topicTitle must be empty when topicPageID is null", i+1)
			}
		} else {
			title, ok := allowedTopics[*question.TopicPageID]
			if !ok {
				return fmt.Errorf("question %d: topicPageID %d is not allowed", i+1, *question.TopicPageID)
			}
			if question.TopicTitle != title {
				return fmt.Errorf("question %d: topicTitle does not match topicPageID %d", i+1, *question.TopicPageID)
			}
		}

		switch question.KnowledgeBasis {
		case "course":
			if len(question.Sources) != 0 {
				return fmt.Errorf("question %d: course knowledge must not have sources", i+1)
			}
		case "external_knowledge", "mixed":
			for sourceIndex, source := range question.Sources {
				if strings.TrimSpace(source.Title) == "" || strings.TrimSpace(source.URL) == "" {
					return fmt.Errorf("question %d: source %d must have title and url", i+1, sourceIndex+1)
				}
			}
		default:
			return fmt.Errorf("question %d: invalid knowledgeBasis %q", i+1, question.KnowledgeBasis)
		}
	}

	return nil
}

func validateGeneratedFeedback(feedback ports.GeneratedFeedback, in ports.GenerateFeedbackRequest) error {
	allowedTopics := make(map[int64]struct{}, len(in.AllowedTopics))
	for _, topic := range in.AllowedTopics {
		allowedTopics[topic.PageID] = struct{}{}
	}

	validateTopic := func(item ports.GeneratedTopicFeedback, field string) error {
		if item.TopicPageID == nil {
			return nil
		}
		if _, ok := allowedTopics[*item.TopicPageID]; !ok {
			return fmt.Errorf("%s contains unknown topicPageID %d", field, *item.TopicPageID)
		}
		return nil
	}

	for _, item := range feedback.MasteredTopics {
		if err := validateTopic(item, "masteredTopics"); err != nil {
			return err
		}
	}
	for _, item := range feedback.TopicsToReview {
		if err := validateTopic(item, "topicsToReview"); err != nil {
			return err
		}
	}
	for _, item := range feedback.NextSteps {
		if item.TopicPageID != nil {
			if _, ok := allowedTopics[*item.TopicPageID]; !ok {
				return fmt.Errorf("nextSteps contains unknown topicPageID %d", *item.TopicPageID)
			}
		}
	}

	return nil
}

func buildRoundPrompt(in ports.GenerateRoundRequest) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Generate exactly %d multiple-choice questions for an adaptive test.\n\n", in.QuestionCount)
	fmt.Fprintf(&b, "COURSE: %s\n", in.CourseTitle)
	fmt.Fprintf(&b, "CURRENT TOPIC: %s\n", in.TopicTitle)
	fmt.Fprintf(&b, "ROUND: %d\n", in.RoundNumber)
	fmt.Fprintf(&b, "STRATEGY: %s\n\n", in.Strategy)

	writeCourseMaterials(&b, in.Materials)
	writeOrdinaryTests(&b, in.OrdinaryTests)
	writePreviousResults(&b, in.PreviousResults)
	writeAllowedTopics(&b, in.AllowedTopics)

	b.WriteString(roundOutputContract)
	return b.String()
}

const roundOutputContract = "OUTPUT CONTRACT\n\n" +
	"Return exactly one JSON object with exactly these top-level fields: \n" +
	"{ \"title\": \"string\", \"instructions\": \"string\", \"questions\": [] }\n\n" +
	"The questions array must contain exactly the requested number of questions.\n\n" +
	"Each question object must contain exactly: topicPageID, topicTitle, question, difficulty, options, correctOptionKey, explanation, knowledgeBasis, sources.\n\n" +
	"Rules:\n" +
	"1. topicPageID must be one of ALLOWED TOPICS, or null only when no allowed topic is appropriate.\n" +
	"2. topicTitle must correspond to topicPageID. If topicPageID is null, use an empty string.\n" +
	"3. question must be one clear educational multiple-choice question.\n" +
	"4. difficulty must be a JSON integer from 1 to 5: 1 very easy, 2 easy, 3 medium, 4 hard, 5 very hard. Never use a string.\n" +
	"5. options must contain exactly four objects with exactly key and text fields. Keys must be A, B, C, D.\n" +
	"6. correctOptionKey must be exactly A, B, C, or D and identify the correct option.\n" +
	"7. explanation must briefly explain why the correct answer is correct.\n" +
	"8. knowledgeBasis must be exactly course, external_knowledge, or mixed. Prefer course knowledge when sufficient.\n" +
	"9. sources must contain objects with exactly title and url. If knowledgeBasis is course, sources must be []. Never invent URLs.\n" +
	"10. Do not copy ordinary-test questions verbatim. Use previous results and STRATEGY to adapt difficulty and topics.\n\n" +
	"JSON rules: return valid JSON only; no Markdown, code fences, comments, or extra fields; numbers are JSON numbers, not strings; null is JSON null."

func writeCourseMaterials(w io.Writer, materials []ports.MaterialContext) {
	fmt.Fprint(w, "COURSE MATERIALS:\n")
	if len(materials) == 0 {
		fmt.Fprint(w, "(none)\n\n")
		return
	}
	for _, material := range materials {
		fmt.Fprintf(w, "[page_id=%d] %s\n%s\n\n", material.PageID, material.Title, material.Content)
	}
}

func writeOrdinaryTests(w io.Writer, tests []ports.OrdinaryTestContext) {
	fmt.Fprint(w, "ORDINARY TESTS:\n")
	if len(tests) == 0 {
		fmt.Fprint(w, "(none)\n\n")
		return
	}
	for _, test := range tests {
		fmt.Fprintf(w, "[test_id=%d] %s\n", test.TestID, test.Title)
		for _, question := range test.Questions {
			fmt.Fprintf(w, "Q: %s\nOptions: %s\n\n", question.Question, strings.Join(question.Options, " | "))
		}
	}
}

func writePreviousResults(w io.Writer, results []ports.PreviousResultContext) {
	fmt.Fprint(w, "PREVIOUS RESULTS:\n")
	if len(results) == 0 {
		fmt.Fprint(w, "(none)\n\n")
		return
	}
	for _, result := range results {
		topicID := int64(0)
		if result.TopicPageID != nil {
			topicID = *result.TopicPageID
		}
		fmt.Fprintf(w, "topic=%d %s: %d/%d correct\n", topicID, result.TopicTitle, result.CorrectCount, result.TotalCount)
	}
	fmt.Fprint(w, "\n")
}

func writeAllowedTopics(w io.Writer, topics []ports.AllowedTopic) {
	fmt.Fprint(w, "ALLOWED TOPICS:\n")
	if len(topics) == 0 {
		fmt.Fprint(w, "(none)\n\n")
		return
	}
	for _, topic := range topics {
		fmt.Fprintf(w, "[page_id=%d] %s\n", topic.PageID, topic.Title)
	}
	fmt.Fprint(w, "\n")
}

func buildFeedbackPrompt(in ports.GenerateFeedbackRequest) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Generate concise learning feedback for course %s.\n\n", in.CourseTitle)
	b.WriteString("ROUND RESULTS:\n")

	if len(in.RoundResults) == 0 {
		b.WriteString("(none)\n\n")
	} else {
		for _, result := range in.RoundResults {
			fmt.Fprintf(&b, "Round %d: %d/%d correct\n", result.RoundNumber, result.CorrectCount, result.TotalCount)
			for _, topic := range result.Topics {
				topicID := int64(0)
				if topic.TopicPageID != nil {
					topicID = *topic.TopicPageID
				}
				fmt.Fprintf(&b, "topic=%d %s: %d/%d correct\n", topicID, topic.TopicTitle, topic.CorrectCount, topic.TotalCount)
			}
			b.WriteString("\n")
		}
	}

	writeAllowedTopics(&b, in.AllowedTopics)
	b.WriteString(feedbackOutputContract)
	return b.String()
}

const feedbackOutputContract = "OUTPUT CONTRACT\n\n" +
	"Return exactly one JSON object with exactly these fields: summary, masteredTopics, topicsToReview, nextSteps.\n" +
	"masteredTopics and topicsToReview items contain exactly title, topicPageID, reason.\n" +
	"nextSteps items contain exactly title, description, topicPageID.\n\n" +
	"Rules:\n" +
	"1. Use only topicPageID values from ALLOWED TOPICS, or null.\n" +
	"2. Do not invent topics.\n" +
	"3. Base feedback only on ROUND RESULTS.\n" +
	"4. masteredTopics are topics with strong demonstrated performance.\n" +
	"5. topicsToReview are topics with mistakes or insufficient evidence of mastery.\n" +
	"6. nextSteps are concrete study actions connected to allowed topics.\n" +
	"7. Keep feedback concise and useful.\n\n" +
	"JSON rules: return valid JSON only; no Markdown, code fences, comments, or extra fields; numbers are JSON numbers, not strings; null is JSON null."

var _ ports.LLMClient = (*Client)(nil)
