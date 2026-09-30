package ports

import "context"

type LLMClient interface {
	GenerateRound(ctx context.Context, request GenerateRoundRequest) (GeneratedRound, GenerationMetadata, error)
	GenerateFeedback(ctx context.Context, request GenerateFeedbackRequest) (GeneratedFeedback, error)
}

type GenerateRoundRequest struct {
	CourseTitle     string
	TopicTitle      string
	Materials       []MaterialContext
	OrdinaryTests   []OrdinaryTestContext
	PreviousResults []PreviousResultContext
	QuestionCount   int
	RoundNumber     int
	Strategy        string
	AllowedTopics   []AllowedTopic
}

type MaterialContext struct {
	PageID  int64
	Title   string
	Content string
}

type OrdinaryTestContext struct {
	TestID    int64
	Title     string
	Questions []OrdinaryQuestionContext
}

type OrdinaryQuestionContext struct {
	Question       string
	Options        []string
	SelectedAnswer string
	IsCorrect      bool
}

type PreviousResultContext struct {
	TopicPageID  *int64
	TopicTitle   string
	CorrectCount int
	TotalCount   int
}

type AllowedTopic struct {
	PageID int64
	Title  string
}

type GeneratedRound struct {
	Title        string
	Instructions string
	Questions    []GeneratedQuestion
}

type GeneratedQuestion struct {
	TopicPageID      *int64
	TopicTitle       string
	Question         string
	Difficulty       int
	Options          []GeneratedOption
	CorrectOptionKey string
	Explanation      string
	KnowledgeBasis   string
	Sources          []GeneratedSource
}

type GeneratedOption struct {
	Key  string
	Text string
}

type GeneratedSource struct {
	Title string
	URL   string
}

type GenerationMetadata struct {
	Model         string
	PromptVersion string
	PromptHash    string
	InputTokens   *int
	OutputTokens  *int
}

type GenerateFeedbackRequest struct {
	CourseTitle   string
	Topics        []FeedbackTopicContext
	RoundResults  []RoundResultContext
	AllowedTopics []AllowedTopic
}

type FeedbackTopicContext struct {
	PageID int64
	Title  string
}

type RoundResultContext struct {
	RoundNumber  int
	CorrectCount int
	TotalCount   int
	Topics       []TopicResultContext
}

type TopicResultContext struct {
	TopicPageID  *int64
	TopicTitle   string
	CorrectCount int
	TotalCount   int
}

type GeneratedFeedback struct {
	Summary        string
	MasteredTopics []GeneratedTopicFeedback
	TopicsToReview []GeneratedTopicFeedback
	NextSteps      []GeneratedNextStep
}

type GeneratedTopicFeedback struct {
	Title       string
	TopicPageID *int64
	Reason      string
}

type GeneratedNextStep struct {
	Title       string
	Description string
	TopicPageID *int64
}
