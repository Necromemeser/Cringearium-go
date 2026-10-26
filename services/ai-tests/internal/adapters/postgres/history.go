package postgres

import (
	"context"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
)

// ListCompletedSessions scopes history on the server; callers cannot request another user's evidence.
func (r *Repository) ListCompletedSessions(ctx context.Context, userID, courseID int64, topicPageID *int64, since time.Time, limit int) ([]domain.AdaptiveSession, error) {
	if limit <= 0 {
		return nil, nil
	}
	var ids []string
	err := r.db.SelectContext(ctx, &ids, `SELECT id FROM adaptive_sessions
 WHERE user_id=$1 AND course_id=$2 AND status='completed' AND completed_at >= $4
 AND ($3::bigint IS NULL OR topic_page_id=$3)
 ORDER BY completed_at DESC,id DESC LIMIT $5`, userID, courseID, topicPageID, since, limit)
	if err != nil {
		return nil, err
	}
	sessions := make([]domain.AdaptiveSession, 0, len(ids))
	for _, id := range ids {
		session, err := r.GetSession(ctx, id, userID)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}
