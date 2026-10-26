package ports

import (
	"context"
	"errors"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
)

// ErrNotFound means the requested session does not exist or is not owned by the user.
var ErrNotFound = errors.New("session not found")

type SessionRepository interface {
	ListCompletedSessions(ctx context.Context, userID, courseID int64, topicPageID *int64, since time.Time, limit int) ([]domain.AdaptiveSession, error)

	CreateSession(
		ctx context.Context,
		session domain.AdaptiveSession,
	) error

	GetSession(
		ctx context.Context,
		sessionID string,
		userID int64,
	) (domain.AdaptiveSession, error)

	GetAdminStats(ctx context.Context) (domain.AdminAIStats, error)

	SaveRound(
		ctx context.Context,
		round domain.AdaptiveRound,
	) (domain.AdaptiveRound, error)

	CompleteSession(
		ctx context.Context,
		sessionID string,
		feedback domain.AdaptiveFeedback,
	) error

	FailSession(
		ctx context.Context,
		sessionID string,
	) error
}
