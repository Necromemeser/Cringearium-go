package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/ports"
)

type repositoryMock struct {
	conversation domain.Conversation
	messages     []domain.Message
	saved        []domain.Message
	getErr       error
	usage        int
	usageErr     error
}

func (m *repositoryMock) CreateConversation(context.Context, domain.Conversation) (domain.Conversation, error) {
	return m.conversation, nil
}

func (m *repositoryMock) GetConversation(_ context.Context, id, userID int64) (domain.Conversation, error) {
	if m.getErr != nil {
		return domain.Conversation{}, m.getErr
	}
	if m.conversation.ID != id || m.conversation.UserID != userID {
		return domain.Conversation{}, ports.ErrNotFound
	}
	return m.conversation, nil
}

func (m *repositoryMock) ListConversations(context.Context, int64) ([]domain.Conversation, error) {
	return nil, nil
}

func (m *repositoryMock) DeleteConversation(context.Context, int64, int64) error {
	return nil
}

func (m *repositoryMock) ListMessages(context.Context, int64, int64) ([]domain.Message, error) {
	return append([]domain.Message(nil), m.messages...), nil
}

func (m *repositoryMock) SaveMessage(_ context.Context, message domain.Message) (domain.Message, error) {
	message.ID = int64(len(m.saved) + 1)
	message.CreatedAt = time.Now()
	m.saved = append(m.saved, message)
	m.messages = append(m.messages, message)
	return message, nil
}

func (m *repositoryMock) UpdateConversationName(context.Context, int64, int64, string) error {
	return nil
}
func (m *repositoryMock) ReserveAIRequest(context.Context, int64, int) (int, error) {
	if m.usageErr != nil {
		return 0, m.usageErr
	}
	m.usage++
	return m.usage, nil
}

func (m *repositoryMock) GetAdminStats(context.Context) (domain.AdminStats, error) {
	return domain.AdminStats{}, nil
}

type llmMock struct {
	messages []ports.LLMMessage
}

func (m *llmMock) Complete(context.Context, []ports.LLMMessage) (string, error) {
	return "Учебный чат", nil
}

func (m *llmMock) Stream(_ context.Context, messages []ports.LLMMessage, onChunk func(string) error) (string, error) {
	m.messages = messages
	if err := onChunk("Привет"); err != nil {
		return "", err
	}
	if err := onChunk("!"); err != nil {
		return "", err
	}
	return "Привет!", nil
}

func TestStreamResponseSavesUserAndAIMessage(t *testing.T) {
	userID := int64(42)
	repo := &repositoryMock{
		conversation: domain.Conversation{ID: 7, UserID: userID, Name: "Тест"},
		messages: []domain.Message{
			{ID: 1, ConversationID: 7, UserID: &userID, Content: "Старый вопрос"},
			{ID: 2, ConversationID: 7, Content: "Старый ответ", IsAIResponse: true},
		},
	}
	llm := &llmMock{}
	service := NewChatService(repo, llm)

	var chunks string
	err := service.StreamResponse(context.Background(), 7, userID, "Новый вопрос", func(chunk string) error {
		chunks += chunk
		return nil
	})
	if err != nil {
		t.Fatalf("StreamResponse() error = %v", err)
	}

	if chunks != "Привет!" {
		t.Fatalf("chunks = %q, want %q", chunks, "Привет!")
	}
	if len(repo.saved) != 2 {
		t.Fatalf("saved messages = %d, want 2", len(repo.saved))
	}
	if repo.saved[0].Content != "Новый вопрос" || repo.saved[0].IsAIResponse {
		t.Fatalf("first saved message = %+v", repo.saved[0])
	}
	if repo.saved[1].Content != "Привет!" || !repo.saved[1].IsAIResponse {
		t.Fatalf("second saved message = %+v", repo.saved[1])
	}
	if len(llm.messages) != 4 {
		t.Fatalf("llm messages = %d, want 4", len(llm.messages))
	}
	if llm.messages[0].Role != "system" {
		t.Fatalf("first llm message role = %q", llm.messages[0].Role)
	}
	if llm.messages[3].Content != "Новый вопрос" {
		t.Fatalf("last llm message = %+v", llm.messages[3])
	}
}

func TestReserveAIRequestRejectsDailyLimit(t *testing.T) {
	repo := &repositoryMock{usageErr: ports.ErrDailyLimitReached}
	service := NewChatService(repo, &llmMock{})

	err := func() error {
		_, err := service.ReserveAIRequest(context.Background(), 42)
		return err
	}()
	if !errors.Is(err, ErrDailyLimitReached) {
		t.Fatalf("error = %v, want ErrDailyLimitReached", err)
	}
}

func TestStreamResponseRejectsForeignConversation(t *testing.T) {
	repo := &repositoryMock{
		conversation: domain.Conversation{ID: 7, UserID: 42},
	}
	service := NewChatService(repo, &llmMock{})

	err := service.StreamResponse(context.Background(), 7, 99, "Привет", func(string) error { return nil })
	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("error = %v, want ErrConversationNotFound", err)
	}
	if len(repo.saved) != 0 {
		t.Fatalf("saved messages = %d, want 0", len(repo.saved))
	}
}

func TestStreamResponseRejectsEmptyMessage(t *testing.T) {
	service := NewChatService(&repositoryMock{}, &llmMock{})
	if !errors.Is(service.StreamResponse(context.Background(), 1, 1, "  ", func(string) error { return nil }), ErrEmptyMessage) {
		t.Fatal("expected ErrEmptyMessage")
	}
}
