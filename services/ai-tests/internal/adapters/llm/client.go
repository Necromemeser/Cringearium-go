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

const promptVersion = "adaptive-v1"

type Client struct {
 logger *slog.Logger
 baseURL string
 apiKey string
 model string
 http *http.Client
}

func NewClient(logger *slog.Logger, baseURL, apiKey, model string) *Client {
 return &Client{logger: logger,baseURL: strings.TrimRight(baseURL, "/"), apiKey: apiKey, model: model, http: &http.Client{Timeout: 90 * time.Second}}
}

func (c *Client) GenerateRound(ctx context.Context, in ports.GenerateRoundRequest) (ports.GeneratedRound, ports.GenerationMetadata, error) {
 prompt := buildRoundPrompt(in)
 content, inputTokens, outputTokens, err := c.complete(ctx, prompt)
 if err != nil { return ports.GeneratedRound{}, ports.GenerationMetadata{}, err }
 var generated ports.GeneratedRound
 if err := decodeJSON(content, &generated); err != nil { return ports.GeneratedRound{}, ports.GenerationMetadata{}, fmt.Errorf("decode generated round: %w", err) }
 sum := sha256.Sum256([]byte(prompt))
 return generated, ports.GenerationMetadata{Model:c.model, PromptVersion:promptVersion, PromptHash:hex.EncodeToString(sum[:]), InputTokens:inputTokens, OutputTokens:outputTokens}, nil
}

func (c *Client) GenerateFeedback(ctx context.Context, in ports.GenerateFeedbackRequest) (ports.GeneratedFeedback, error) {
 prompt := buildFeedbackPrompt(in)
 content, _, _, err := c.complete(ctx, prompt)
 if err != nil { return ports.GeneratedFeedback{}, err }
 var generated ports.GeneratedFeedback
 if err := decodeJSON(content, &generated); err != nil { return ports.GeneratedFeedback{}, fmt.Errorf("decode generated feedback: %w", err) }
 return generated, nil
}

func (c *Client) complete(ctx context.Context, prompt string) (string, *int, *int, error) {
 start := time.Now()
 apiKey := strings.TrimSpace(c.apiKey)
 if apiKey == "" {
  return "", nil, nil, fmt.Errorf("LLM_API_KEY is empty")
 }
 c.logger.Info("sending llm request", "model", c.model, "auth_configured", true)
 payload := map[string]any{"model":c.model,"messages":[]map[string]string{{"role":"system","content":"Generate educational adaptive tests. Return only valid JSON."},{"role":"user","content":prompt}},"thinking":map[string]string{"type":"disabled"},"max_tokens":4096,"temperature":0.2,"response_format":map[string]string{"type":"json_object"}}
 body, err := json.Marshal(payload)
 if err != nil { return "", nil, nil, err }
 req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
 if err != nil { return "", nil, nil, err }
 req.Header.Set("Content-Type","application/json")
 req.Header.Set("Accept","application/json")
 req.Header.Set("Authorization", "Bearer "+apiKey)
 resp, err := c.http.Do(req)
 if err != nil {
  c.logger.Error("llm request failed", "model", c.model, "duration_ms", time.Since(start).Milliseconds(), "error", err)
  return "", nil, nil, fmt.Errorf("llm request: %w", err)
 }
 defer resp.Body.Close()
 data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
 if err != nil { return "", nil, nil, err }
 if resp.StatusCode < 200 || resp.StatusCode >= 300 {
  c.logger.Error("llm returned non-success status", "model", c.model, "status", resp.StatusCode, "duration_ms", time.Since(start).Milliseconds())
  return "", nil, nil, fmt.Errorf("llm returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
 }
 var envelope struct {
 Choices []struct { FinishReason string "json:\"finish_reason\""; Message struct { Content string "json:\"content\"" } "json:\"message\"" } "json:\"choices\""
 Usage struct { PromptTokens int "json:\"prompt_tokens\""; CompletionTokens int "json:\"completion_tokens\"" } "json:\"usage\""
 }
 if err := json.Unmarshal(data, &envelope); err != nil { return "", nil, nil, fmt.Errorf("decode llm response: %w", err) }
 if len(envelope.Choices) == 0 { return "", nil, nil, fmt.Errorf("llm returned no choices") }
 choice := envelope.Choices[0]
 if choice.FinishReason == "length" { return "", nil, nil, fmt.Errorf("llm response truncated: max_tokens reached") }
 if strings.TrimSpace(choice.Message.Content) == "" { return "", nil, nil, fmt.Errorf("llm returned no content") }
 inputTokens := envelope.Usage.PromptTokens
 outputTokens := envelope.Usage.CompletionTokens
 c.logger.Info("llm request completed", "model", c.model, "status", resp.StatusCode, "duration_ms", time.Since(start).Milliseconds(), "finish_reason", choice.FinishReason, "input_tokens", inputTokens, "output_tokens", outputTokens)
 return strings.TrimSpace(choice.Message.Content), &inputTokens, &outputTokens, nil
}

func decodeJSON(content string, target any) error {
 content = strings.TrimSpace(content)
 content = strings.TrimPrefix(content, "```json")
 content = strings.TrimPrefix(content, "```")
 content = strings.TrimSuffix(content, "```")
 return json.Unmarshal([]byte(strings.TrimSpace(content)), target)
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
	"Return exactly one JSON object with exactly these top-level fields: \\n" +
	"{ \\"title\\": \\"string\\", \\"instructions\\": \\"string\\", \\"questions\\": [] }\n\n" +
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

func writeCourseMaterials(b *strings.Builder, materials []ports.MaterialContext) {
	b.WriteString("COURSE MATERIALS:\n")
	if len(materials) == 0 {
		b.WriteString("(none)\n\n")
		return
	}
	for _, material := range materials {
		fmt.Fprintf(&b, "[page_id=%d] %s\n%s\n\n", material.PageID, material.Title, material.Content)
	}
}

func writeOrdinaryTests(b *strings.Builder, tests []ports.OrdinaryTestContext) {
	b.WriteString("ORDINARY TESTS:\n")
	if len(tests) == 0 {
		b.WriteString("(none)\n\n")
		return
	}
	for _, test := range tests {
		fmt.Fprintf(&b, "[test_id=%d] %s\n", test.TestID, test.Title)
		for _, question := range test.Questions {
			fmt.Fprintf(&b, "Q: %s\nOptions: %s\n\n", question.Question, strings.Join(question.Options, " | "))
		}
	}
}

func writePreviousResults(b *strings.Builder, results []ports.PreviousResultContext) {
	b.WriteString("PREVIOUS RESULTS:\n")
	if len(results) == 0 {
		b.WriteString("(none)\n\n")
		return
	}
	for _, result := range results {
		topicID := int64(0)
		if result.TopicPageID != nil {
			topicID = *result.TopicPageID
		}
		fmt.Fprintf(&b, "topic=%d %s: %d/%d correct\n", topicID, result.TopicTitle, result.CorrectCount, result.TotalCount)
	}
	b.WriteString("\n")
}

func writeAllowedTopics(b *strings.Builder, topics []ports.AllowedTopic) {
	b.WriteString("ALLOWED TOPICS:\n")
	if len(topics) == 0 {
		b.WriteString("(none)\n\n")
		return
	}
	for _, topic := range topics {
		fmt.Fprintf(b, "[page_id=%d] %s\n", topic.PageID, topic.Title)
	}
	b.WriteString("\n")
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
