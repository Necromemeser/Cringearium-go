package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/ports"
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrEmptyMessage         = errors.New("input is empty")
	ErrDailyLimitReached    = errors.New("daily ai request limit reached")
)

const maxMessageLength = 32 * 1024
const maxHistoryMessages = 40
const dailyRequestLimit = 10

type ChatService struct {
	repository ports.Repository
	llm        ports.LLMClient
}

func NewChatService(repository ports.Repository, llm ports.LLMClient) *ChatService {
	return &ChatService{repository: repository, llm: llm}
}

func (s *ChatService) CreateConversation(ctx context.Context, userID int64, name string) (domain.Conversation, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Новый чат"
	}
	now := time.Now()
	return s.repository.CreateConversation(ctx, domain.Conversation{
		UserID:    userID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *ChatService) GetConversation(ctx context.Context, id, userID int64) (domain.Conversation, error) {
	item, err := s.repository.GetConversation(ctx, id, userID)
	if errors.Is(err, ports.ErrNotFound) {
		return domain.Conversation{}, ErrConversationNotFound
	}
	return item, err
}

func (s *ChatService) ListConversations(ctx context.Context, userID int64) ([]domain.Conversation, error) {
	return s.repository.ListConversations(ctx, userID)
}

func (s *ChatService) DeleteConversation(ctx context.Context, id, userID int64) error {
	err := s.repository.DeleteConversation(ctx, id, userID)
	if errors.Is(err, ports.ErrNotFound) {
		return ErrConversationNotFound
	}
	return err
}

func (s *ChatService) ListMessages(ctx context.Context, conversationID, userID int64) ([]domain.Message, error) {
	items, err := s.repository.ListMessages(ctx, conversationID, userID)
	if errors.Is(err, ports.ErrNotFound) {
		return nil, ErrConversationNotFound
	}
	return items, err
}

func (s *ChatService) ReserveAIRequest(ctx context.Context, userID int64) (int, error) {
	count, err := s.repository.ReserveAIRequest(ctx, userID, dailyRequestLimit)
	if errors.Is(err, ports.ErrDailyLimitReached) {
		return 0, ErrDailyLimitReached
	}
	return count, err
}

func (s *ChatService) StreamResponse(ctx context.Context, conversationID, userID int64, input string, onChunk func(string) error) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return ErrEmptyMessage
	}
	if len(input) > maxMessageLength {
		return errors.New("message is too long")
	}

	if _, err := s.GetConversation(ctx, conversationID, userID); err != nil {
		return err
	}

	if _, err := s.repository.SaveMessage(ctx, domain.Message{
		ConversationID: conversationID,
		UserID:         &userID,
		Content:        input,
		IsAIResponse:   false,
	}); err != nil {
		return err
	}

	history, err := s.repository.ListMessages(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if len(history) > maxHistoryMessages {
		history = history[len(history)-maxHistoryMessages:]
	}

	messages := make([]ports.LLMMessage, 0, len(history)+1)
	messages = append(messages, ports.LLMMessage{
		Role:    "system",
		Content: systemPrompt,
	})
	for _, message := range history {
		role := "user"
		if message.IsAIResponse {
			role = "assistant"
		}
		messages = append(messages, ports.LLMMessage{Role: role, Content: message.Content})
	}

	fullResponse, err := s.llm.Stream(ctx, messages, onChunk)
	if err != nil {
		return err
	}
	fullResponse = strings.TrimSpace(fullResponse)
	if fullResponse == "" {
		return errors.New("llm returned empty response")
	}

	_, err = s.repository.SaveMessage(ctx, domain.Message{
		ConversationID: conversationID,
		Content:        fullResponse,
		IsAIResponse:   true,
	})
	return err
}

func (s *ChatService) GetAdminStats(ctx context.Context) (domain.AdminStats, error) {
	return s.repository.GetAdminStats(ctx)
}
