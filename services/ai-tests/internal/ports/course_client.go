package ports

import (
	"context"
	"errors"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
)

var ErrCourseUnavailable = errors.New("course is unavailable")

var ErrInvalidCourseContext = errors.New("invalid course practice context")

type CourseClient interface {
	GetCourseContext(
		ctx context.Context,
		userID int64,
		courseID int64,
		topicPageID *int64,
	) (CourseContext, error)
}

type CourseContext struct {
	Skills        []domain.LearningSkill
	CourseID      int64
	CourseTitle   string
	TopicTitle    string
	Materials     []CourseMaterial
	OrdinaryTests []OrdinaryTest
	TestResults   []OrdinaryTestResult
	AllowedTopics []CourseTopic
}

type CourseMaterial struct {
	PageID  int64
	Title   string
	Content string
}

type OrdinaryTest struct {
	ID        int64
	PageID    int64
	Title     string
	Questions []OrdinaryTestQuestion
}

type OrdinaryTestQuestion struct {
	SelectedAnswer string
	CorrectAnswer  string
	Answered       bool
	IsCorrect      bool
	Question       string
	Options        []string
}

type OrdinaryTestResult struct {
	TestID       int64
	PageID       int64
	Title        string
	CorrectCount int
	TotalCount   int
}

type CourseTopic struct {
	PageID int64
	Title  string
}
