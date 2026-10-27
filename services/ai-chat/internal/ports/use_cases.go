package ports

import (
	"context"

	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/domain"
)

// ChatServiceUseCases is the input port used by transport adapters.
type ChatServiceUseCases interface {
	CreateConversation(ctx context.Context, userID int64, name string) (domain.Conversation, error)
	GetConversation(ctx context.Context, id, userID int64) (domain.Conversation, error)
	ListConversations(ctx context.Context, userID int64) ([]domain.Conversation, error)
	DeleteConversation(ctx context.Context, id, userID int64) error
	ListMessages(ctx context.Context, conversationID, userID int64) ([]domain.Message, error)
	ReserveAIRequest(ctx context.Context, userID int64) (int, error)
	StreamResponse(ctx context.Context, conversationID, userID int64, input string, onChunk func(string) error) error
	GetAdminStats(ctx context.Context) (domain.AdminStats, error)
}
