package ports

import (
	"context"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
)

type SessionRepository interface {
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

	SaveAnswers(
		ctx context.Context,
		sessionID string,
		roundID int64,
		answers []domain.AdaptiveAnswer,
	) error

	CompleteRound(
		ctx context.Context,
		sessionID string,
		roundID int64,
	) error

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
