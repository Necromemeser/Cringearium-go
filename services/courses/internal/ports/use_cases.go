package ports

import (
	"context"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

// CoursesUseCases is the input port used by transport adapters.
type CoursesUseCases interface {
	GetAll(ctx context.Context) ([]*domain.Course, error)
	GetAdminCourses(ctx context.Context) ([]domain.AdminCourse, error)
	GetAssessmentResults(ctx context.Context) ([]domain.AssessmentResult, error)
	GetByID(ctx context.Context, id int64) (*domain.CourseDetails, error)
	FindEnrolled(ctx context.Context, userID int64) ([]*domain.Course, error)
	Enroll(ctx context.Context, userID, courseID int64) error
	HasAccess(ctx context.Context, userID, courseID int64) (bool, error)
	GetCompletedPages(ctx context.Context, userID, courseID int64) ([]int64, error)
	CompletePage(ctx context.Context, userID, pageID int64) error
	GetAIContext(
		ctx context.Context,
		userID int64,
		courseID int64,
		topicPageID *int64,
	) (*domain.AIContext, error)
	GetTest(ctx context.Context, userID, pageID int64) (*domain.Test, *domain.TestAttempt, error)
	SubmitTest(ctx context.Context, userID, testID int64, answers []domain.TestAttemptAnswer) (*domain.TestResult, error)
}
