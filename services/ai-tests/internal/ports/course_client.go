package ports

import "context"

type CourseClient interface {
	GetCourseContext(
		ctx context.Context,
		userID int64,
		courseID int64,
		topicPageID *int64,
	) (CourseContext, error)
}

type CourseContext struct {
	CourseID      int64
	CourseTitle   string
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
	Title     string
	Questions []OrdinaryTestQuestion
}

type OrdinaryTestQuestion struct {
	Question string
	Options  []string
}

type OrdinaryTestResult struct {
	TestID       int64
	CorrectCount int
	TotalCount   int
}

type CourseTopic struct {
	PageID int64
	Title  string
}
