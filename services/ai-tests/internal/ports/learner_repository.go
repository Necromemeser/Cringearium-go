package ports

import (
	"context"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
)

type LearnerRepository interface {
	GetSkillStates(ctx context.Context, userID int64, codes []string) ([]domain.SkillState, error)
	// The bool reports a new completion, so concurrent retries do not generate a second round twice.
	// Answers, round completion and novel learning evidence must commit in one transaction.
	SubmitLearningRound(ctx context.Context, userID int64, sessionID string, roundID int64, answers []domain.AdaptiveAnswer) (bool, error)
}
