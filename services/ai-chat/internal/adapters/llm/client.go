package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/ports"
)

const requestTimeout = 90 * time.Second

type Client struct {
	log     *slog.Logger
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

func NewClient(log *slog.Logger, baseURL, apiKey, model string) *Client {
	return &Client{
		log: log, baseURL: strings.TrimRight(baseURL, "/"), apiKey: apiKey, model: model,
		http: &http.Client{Timeout: requestTimeout},
	}
}

func (c *Client) Stream(ctx context.Context, messages []ports.LLMMessage, onChunk func(string) error) (string, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return "", fmt.Errorf("LLM_API_KEY is empty")
	}

	bodyMessages := make([]map[string]string, 0, len(messages))
	for _, message := range messages {
		bodyMessages = append(bodyMessages, map[string]string{"role": message.Role, "content": message.Content})
	}

	payload := map[string]any{
		"model": c.model,
		"messages": bodyMessages,
		"stream": true,
		"thinking": map[string]string{"type": "disabled"},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal llm request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create llm request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return "", fmt.Errorf("llm returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	var full strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 4<<20)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line == "data: [DONE]" {
			continue
		}
		if strings.HasPrefix(line, "data:") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			c.log.Warn("failed to parse llm stream chunk", "error", err)
			continue
		}
		if len(chunk.Choices) == 0 || chunk.Choices[0].Delta.Content == "" {
			continue
		}

		content := chunk.Choices[0].Delta.Content
		full.WriteString(content)
		if err := onChunk(content); err != nil {
			return "", err
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read llm stream: %w", err)
	}

	return full.String(), nil
}

func (c *Client) Complete(ctx context.Context, messages []ports.LLMMessage) (string, error) {
	if strings.TrimSpace(c.apiKey) == "" { return "", fmt.Errorf("LLM_API_KEY is empty") }
	bodyMessages := make([]map[string]string, 0, len(messages))
	for _, message := range messages { bodyMessages = append(bodyMessages, map[string]string{"role": message.Role, "content": message.Content}) }
	payload := map[string]any{"model": c.model, "messages": bodyMessages, "stream": false, "thinking": map[string]string{"type": "disabled"}, "max_tokens": 32}
	body, err := json.Marshal(payload); if err != nil { return "", fmt.Errorf("marshal llm request: %w", err) }
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body)); if err != nil { return "", fmt.Errorf("create llm request: %w", err) }
	req.Header.Set("Content-Type", "application/json"); req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.http.Do(req); if err != nil { return "", fmt.Errorf("llm request: %w", err) }; defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)); return "", fmt.Errorf("llm returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(data))) }
	var result struct { Choices []struct { Message struct { Content string ``json:"content"` } ``json:"message"` } ``json:"choices"` }
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { return "", fmt.Errorf("decode llm response: %w", err) }
	if len(result.Choices) == 0 { return "", errors.New("llm returned no choices") }
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}
var _ ports.LLMClient = (*Client)(nil)

