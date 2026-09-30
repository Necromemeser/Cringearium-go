package domain

import "time"

type KnowledgeBasis string

const (
	BasisCourse            KnowledgeBasis = "course"
	BasisExternalKnowledge KnowledgeBasis = "external_knowledge"
	BasisMixed             KnowledgeBasis = "mixed"
)

type AdaptiveQuestion struct {
	ID          int64
	RoundID     int64
	Position    int
	TopicPageID *int64
	TopicTitle  string

	Question         string
	Difficulty       int
	CorrectOptionKey string
	Explanation      string
	KnowledgeBasis   KnowledgeBasis
	Sources          []QuestionSource

	Options []AdaptiveOption
}

type AdaptiveOption struct {
	ID         int64
	QuestionID int64
	Key        string
	Text       string
	Position   int
}

type AdaptiveAnswer struct {
	SessionID         string
	RoundID           int64
	QuestionID        int64
	SelectedOptionKey string
	IsCorrect         bool
	AnsweredAt        time.Time
}

type QuestionSource struct {
	Title string
	URL   string
}
