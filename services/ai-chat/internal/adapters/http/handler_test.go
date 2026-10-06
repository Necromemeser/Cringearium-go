package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/ports"
)

type handlerRepositoryMock struct {
	conversation domain.Conversation
}

func (m *handlerRepositoryMock) CreateConversation(context.Context, domain.Conversation) (domain.Conversation, error) {
	return m.conversation, nil
}
func (m *handlerRepositoryMock) GetConversation(_ context.Context, id, userID int64) (domain.Conversation, error) {
	if id != m.conversation.ID || userID != m.conversation.UserID {
		return domain.Conversation{}, ports.ErrNotFound
	}
	return m.conversation, nil
}
func (m *handlerRepositoryMock) ListConversations(context.Context, int64) ([]domain.Conversation, error) {
	return []domain.Conversation{m.conversation}, nil
}
func (m *handlerRepositoryMock) DeleteConversation(context.Context, int64, int64) error {
	return nil
}
func (m *handlerRepositoryMock) ListMessages(context.Context, int64, int64) ([]domain.Message, error) {
	return nil, nil
}
func (m *handlerRepositoryMock) SaveMessage(context.Context, domain.Message) (domain.Message, error) {
	return domain.Message{}, nil
}
func (m *handlerRepositoryMock) GetAdminStats(context.Context) (domain.AdminStats, error) {
	return domain.AdminStats{}, nil
}

type handlerLLMMock struct{}

func (handlerLLMMock) Stream(_ context.Context, _ []ports.LLMMessage, onChunk func(string) error) (string, error) {
	if err := onChunk("ответ"); err != nil {
		return "", err
	}
	return "ответ", nil
}

func TestSendMessageStreamsSSE(t *testing.T) {
	service := application.NewChatService(
		&handlerRepositoryMock{conversation: domain.Conversation{ID: 7, UserID: 42}},
		handlerLLMMock{},
	)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/api/chats/7/messages", strings.NewReader(`{"content":"привет"}`))
	req.SetPathValue("id", "7")
	req.Header.Set("X-User-ID", "42")
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.SendMessage(response, req)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if got := response.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("content type = %q", got)
	}
	body := response.Body.String()
	if !strings.Contains(body, `"type":"chunk"`) || !strings.Contains(body, `"content":"ответ"`) {
		t.Fatalf("missing chunk event: %q", body)
	}
	if !strings.Contains(body, `"type":"done"`) {
		t.Fatalf("missing done event: %q", body)
	}
}

func TestSendMessageRejectsForeignConversation(t *testing.T) {
	service := application.NewChatService(
		&handlerRepositoryMock{conversation: domain.Conversation{ID: 7, UserID: 42}},
		handlerLLMMock{},
	)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/api/chats/7/messages", strings.NewReader(`{"content":"привет"}`))
	req.SetPathValue("id", "7")
	req.Header.Set("X-User-ID", "99")
	response := httptest.NewRecorder()

	handler.SendMessage(response, req)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestSendMessageRequiresUserID(t *testing.T) {
	service := application.NewChatService(
		&handlerRepositoryMock{conversation: domain.Conversation{ID: 7, UserID: 42}},
		handlerLLMMock{},
	)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/api/chats/7/messages", strings.NewReader(`{"content":"привет"}`))
	req.SetPathValue("id", "7")
	response := httptest.NewRecorder()

	handler.SendMessage(response, req)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

var _ = errors.Is
