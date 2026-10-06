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
const maxConversationTitleLength = 50

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
	if err != nil {
		return err
	}

	if len(history) == 1 && !history[0].IsAIResponse {
		s.generateConversationTitle(ctx, conversationID, userID, input, fullResponse)
	}

	return nil
}


func (s *ChatService) generateConversationTitle(ctx context.Context, conversationID, userID int64, userMessage, aiResponse string) {
	messages := []ports.LLMMessage{
		{Role: "system", Content: `Придумай короткое название для учебного чата. Название должно отражать тему разговора, а не повторять сообщение пользователя. Правила: 2–6 слов, максимум 50 символов, без кавычек, markdown и точки в конце. Используй язык разговора. Не используй шаблоны вроде «Чат о...», «Разговор о...» или «Помощь с...». Верни только название, без пояснений.`},
		{Role: "user", Content: "Сообщение студента:\n" + userMessage + "\n\nОтвет Кринжика:\n" + aiResponse},
	}
	title, err := s.llm.Complete(ctx, messages)
	if err != nil { return }
	title = strings.TrimSpace(strings.Trim(title, `"`))
	title = strings.TrimRight(title, ".!?;:")
	if title == "" { return }
	if len([]rune(title)) > maxConversationTitleLength { title = string([]rune(title)[:maxConversationTitleLength]) }
	if err := s.repository.UpdateConversationName(ctx, conversationID, userID, title); err != nil { return }
}
func (s *ChatService) GetAdminStats(ctx context.Context) (domain.AdminStats, error) {
	return s.repository.GetAdminStats(ctx)
}
