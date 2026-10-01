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
 fmt.Fprintf(&b, "Generate exactly %d multiple-choice questions. Course: %s. Round: %d. Strategy: %s.\\n", in.QuestionCount, in.CourseTitle, in.RoundNumber, in.Strategy)
 b.WriteString("COURSE MATERIALS:\\n")
 for _, m := range in.Materials { fmt.Fprintf(&b, "[page_id=%d] %s\\n%s\\n", m.PageID, m.Title, m.Content) }
 b.WriteString("ORDINARY TESTS:\\n")
 for _, t := range in.OrdinaryTests { fmt.Fprintf(&b, "[test_id=%d] %s\\n", t.TestID, t.Title); for _, q := range t.Questions { fmt.Fprintf(&b, "Q: %s\\nOptions: %s\\n", q.Question, strings.Join(q.Options, " | ")) } }
 b.WriteString("PREVIOUS RESULTS:\\n")
 for _, r := range in.PreviousResults { id:=int64(0); if r.TopicPageID != nil { id=*r.TopicPageID }; fmt.Fprintf(&b, "topic=%d %s: %d/%d\\n", id, r.TopicTitle, r.CorrectCount, r.TotalCount) }
 b.WriteString("ALLOWED TOPICS:\\n")
 for _, t := range in.AllowedTopics { fmt.Fprintf(&b, "[page_id=%d] %s\\n", t.PageID, t.Title) }
 b.WriteString("Return only JSON with title, instructions, questions. Each question needs topicPageID, topicTitle, question, difficulty, options, correctOptionKey, explanation, knowledgeBasis, sources. topicPageID must be an allowed ID or null. knowledgeBasis must be course, external_knowledge, or mixed. Do not invent source URLs.")
 return b.String()
}

func buildFeedbackPrompt(in ports.GenerateFeedbackRequest) string {
 var b strings.Builder
 fmt.Fprintf(&b, "Generate concise learning feedback for course %s.\\nROUND RESULTS:\\n", in.CourseTitle)
 for _, r := range in.RoundResults { fmt.Fprintf(&b, "Round %d: %d/%d correct\\n", r.RoundNumber, r.CorrectCount, r.TotalCount); for _, t := range r.Topics { id:=int64(0); if t.TopicPageID != nil { id=*t.TopicPageID }; fmt.Fprintf(&b, "topic=%d %s: %d/%d\\n", id, t.TopicTitle, t.CorrectCount, t.TotalCount) } }
 b.WriteString("ALLOWED TOPICS:\\n")
 for _, t := range in.AllowedTopics { fmt.Fprintf(&b, "[page_id=%d] %s\\n", t.PageID, t.Title) }
 b.WriteString("Return only JSON with summary, masteredTopics, topicsToReview, nextSteps. Use only allowed topic IDs.")
 return b.String()
}

var _ ports.LLMClient = (*Client)(nil)