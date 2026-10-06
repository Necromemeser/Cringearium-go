package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/ports"
)

type Repository struct{ db *DB }
func NewRepository(db *DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateConversation(ctx context.Context, conversation domain.Conversation) (domain.Conversation, error) {
	var result domain.Conversation
	err := r.db.conn.QueryRowxContext(ctx, `
		INSERT INTO conversations (user_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, name, created_at, updated_at
	`, conversation.UserID, conversation.Name, conversation.CreatedAt, conversation.UpdatedAt).StructScan(&result)
	return result, err
}
func (r *Repository) GetConversation(ctx context.Context, id, userID int64) (domain.Conversation, error) {
	var result domain.Conversation
	err := r.db.conn.QueryRowxContext(ctx, `
		SELECT id, user_id, name, created_at, updated_at FROM conversations
		WHERE id = $1 AND user_id = $2
	`, id, userID).StructScan(&result)
	if errors.Is(err, sql.ErrNoRows) { return domain.Conversation{}, ports.ErrNotFound }
	return result, err
}
func (r *Repository) ListConversations(ctx context.Context, userID int64) ([]domain.Conversation, error) {
	var rows []domain.Conversation
	err := r.db.conn.SelectContext(ctx, &rows, `
		SELECT id, user_id, name, created_at, updated_at FROM conversations
		WHERE user_id = $1 ORDER BY updated_at DESC, id DESC
	`, userID)
	return rows, err
}
func (r *Repository) DeleteConversation(ctx context.Context, id, userID int64) error {
	result, err := r.db.conn.ExecContext(ctx, `DELETE FROM conversations WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil { return err }
	count, err := result.RowsAffected(); if err != nil { return err }
	if count == 0 { return ports.ErrNotFound }
	return nil
}
func (r *Repository) ListMessages(ctx context.Context, conversationID, userID int64) ([]domain.Message, error) {
	if _, err := r.GetConversation(ctx, conversationID, userID); err != nil { return nil, err }
	var rows []domain.Message
	err := r.db.conn.SelectContext(ctx, &rows, `
		SELECT id, conversation_id, user_id, content, is_ai_response, created_at
		FROM messages WHERE conversation_id = $1 ORDER BY created_at ASC, id ASC
	`, conversationID)
	return rows, err
}
func (r *Repository) SaveMessage(ctx context.Context, message domain.Message) (domain.Message, error) {
	var result domain.Message
	err := r.db.conn.QueryRowxContext(ctx, `
		INSERT INTO messages (conversation_id, user_id, content, is_ai_response)
		VALUES ($1, $2, $3, $4)
		RETURNING id, conversation_id, user_id, content, is_ai_response, created_at
	`, message.ConversationID, message.UserID, message.Content, message.IsAIResponse).StructScan(&result)
	if err != nil { return domain.Message{}, err }
	_, err = r.db.conn.ExecContext(ctx, `UPDATE conversations SET updated_at = $1 WHERE id = $2`, result.CreatedAt, result.ConversationID)
	return result, err
}
func (r *Repository) GetAdminStats(ctx context.Context) (domain.AdminStats, error) {
	var stats domain.AdminStats
	queries := []struct{ target *int64; query string }{
		{&stats.Conversations, `SELECT COUNT(*) FROM conversations`},
		{&stats.Messages, `SELECT COUNT(*) FROM messages`},
		{&stats.UserMessages, `SELECT COUNT(*) FROM messages WHERE is_ai_response = FALSE`},
		{&stats.AIMessages, `SELECT COUNT(*) FROM messages WHERE is_ai_response = TRUE`},
		{&stats.ActiveUsers, `SELECT COUNT(DISTINCT user_id) FROM conversations`},
	}
	for _, item := range queries { if err := r.db.conn.GetContext(ctx, item.target, item.query); err != nil { return domain.AdminStats{}, err } }
	return stats, nil
}
var ErrNotFound = errors.New("conversation not found")
var _ ports.Repository = (*Repository)(nil)
