package ports

import (
	"context"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

type AssessmentRepository interface {
	FindAssessmentResults(ctx context.Context) ([]domain.AssessmentResult, error)
}
