package domain

type AIContext struct {
	CourseID      int64
	CourseTitle   string
	Materials     []AIContextMaterial
	OrdinaryTests []AIContextTest
	TestResults   []AIContextResult
	AllowedTopics []AIContextTopic
}

type AIContextMaterial struct {
	PageID  int64
	Title   string
	Content string
}

type AIContextTest struct {
	ID        int64
	PageID    int64
	Title     string
	Questions []AIContextQuestion
}

type AIContextQuestion struct {
	Question string
	Options  []string
}

type AIContextResult struct {
	TestID       int64
	PageID       int64
	Title        string
	CorrectCount int
	TotalCount   int
}

type AIContextTopic struct {
	PageID int64
	Title  string
}
