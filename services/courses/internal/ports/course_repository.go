package ports

import (
	"context"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

type CourseRepository interface {
	FindAll(ctx context.Context) ([]*domain.Course, error)
	FindByID(ctx context.Context, id int64) (*domain.CourseDetails, error)
	FindEnrolled(ctx context.Context, userID int64) ([]*domain.Course, error)
	HasAccess(ctx context.Context, userID, courseID int64) (bool, error)
	GrantAccess(ctx context.Context, userID, courseID int64) error
	GetCompletedPages(ctx context.Context, userID, courseID int64) ([]int64, error)
	CompletePage(ctx context.Context, userID, pageID int64) error
}
