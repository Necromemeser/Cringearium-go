package ports

import (
	"context"
	"errors"

	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/domain"
)

var ErrNotFound = errors.New("conversation not found")

type Repository interface {
	CreateConversation(ctx context.Context, conversation domain.Conversation) (domain.Conversation, error)
	GetConversation(ctx context.Context, id, userID int64) (domain.Conversation, error)
	ListConversations(ctx context.Context, userID int64) ([]domain.Conversation, error)
	DeleteConversation(ctx context.Context, id, userID int64) error
	ListMessages(ctx context.Context, conversationID, userID int64) ([]domain.Message, error)
	SaveMessage(ctx context.Context, message domain.Message) (domain.Message, error)
	GetAdminStats(ctx context.Context) (domain.AdminStats, error)
}

type LLMClient interface {
	Stream(ctx context.Context, messages []LLMMessage, onChunk func(string) error) (string, error)
}

type LLMMessage struct {
	Role    string
	Content string
}
