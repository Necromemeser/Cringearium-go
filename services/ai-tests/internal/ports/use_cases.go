package ports

import (
	"context"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
)

// AdaptiveTestServiceUseCases is the input port used by transport adapters.
type AdaptiveTestServiceUseCases interface {
	GetLearningProfile(ctx context.Context, userID, courseID int64, topicPageID *int64) (domain.LearningProfile, error)
	GetAdminStats(ctx context.Context) (domain.AdminAIStats, error)
	CreateSession(
		ctx context.Context,
		userID int64,
		courseID int64,
		topicPageID *int64,
		questionCount int,
	) (domain.AdaptiveSession, error)
	GetSession(ctx context.Context, userID int64, sessionID string) (domain.AdaptiveSession, error)
	SubmitAnswers(
		ctx context.Context,
		userID int64,
		sessionID string,
		roundID int64,
		answers map[int64]string,
	) (domain.AdaptiveSession, error)
}
