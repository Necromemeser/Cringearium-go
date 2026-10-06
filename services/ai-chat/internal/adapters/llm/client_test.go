package llm

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/ports"

)

func TestComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/chat/completions" { t.Fatalf("request = %s %s", r.Method, r.URL.Path) }
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" { t.Fatalf("authorization = %q", got) }
		var payload struct { Model string `json:"model"`; Messages []struct { Role string `json:"role"`; Content string `json:"content"` } `json:"messages"`; Stream bool `json:"stream"`; Thinking map[string]string `json:"thinking"`; MaxTokens int `json:"max_tokens"` }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil { t.Fatalf("decode request: %v", err) }
		if payload.Model != "test-model" || payload.Stream || payload.MaxTokens != 32 { t.Fatalf("unexpected payload: %+v", payload) }
		if payload.Thinking["type"] != "disabled" { t.Fatalf("thinking = %+v", payload.Thinking) }
		if len(payload.Messages) != 2 || payload.Messages[0].Role != "system" { t.Fatalf("messages = %+v", payload.Messages) }
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"  Производные  "}}]}`))
	}))
	defer server.Close()
	client := NewClient(slog.Default(), server.URL, "test-key", "test-model")
	got, err := client.Complete(context.Background(), []ports.LLMMessage{{Role:"system", Content:"title"}, {Role:"user", Content:"question"}})
	if err != nil { t.Fatalf("Complete() error = %v", err) }
	if got != "Производные" { t.Fatalf("result = %q", got) }
}

func TestCompleteErrors(t *testing.T) {
	tests := []struct { name string; status int; body string; want string }{
		{"server error", http.StatusBadGateway, "upstream failed", "llm returned status 502"},
		{"invalid json", http.StatusOK, "{", "decode llm response"},
		{"no choices", http.StatusOK, `{"choices":[]}`, "llm returned no choices"},
	}
	for _, tt := range tests { t.Run(tt.name, func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tt.status); _, _ = w.Write([]byte(tt.body)) }))
		defer server.Close()
		client := NewClient(slog.Default(), server.URL, "test-key", "test-model")
		_, err := client.Complete(context.Background(), nil)
		if err == nil || !strings.Contains(err.Error(), tt.want) { t.Fatalf("error = %v, want substring %q", err, tt.want) }
	}) }
}

func TestCompleteRequiresAPIKey(t *testing.T) {
	client := NewClient(slog.Default(), "http://example.com", "", "test-model")
	_, err := client.Complete(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "LLM_API_KEY is empty") { t.Fatalf("error = %v", err) }
}
