package domain

import (
	"encoding/json"
	"time"
)

type Test struct {
	AssessmentType  string
	AssessmentScope string
	ID              int64
	PageID          int64
	PageTitle       string
	PassingScore    int
	Questions       []TestQuestion
}

type TestQuestion struct {
	ID       int64
	TestID   int64
	Question string
	Position int
	Answers  []TestAnswer
}

type TestAnswer struct {
	ID         int64
	QuestionID int64
	Text       string
	Position   int
	IsCorrect  bool
}

type TestAttemptAnswer struct {
	QuestionID      int64 `json:"question_id"`
	AnswerID        int64 `json:"answer_id"`
	CorrectAnswerID int64 `json:"correct_answer_id"`
	IsCorrect       bool  `json:"is_correct"`
}

type TestAttempt struct {
	ID          int64
	Score       int
	Passed      bool
	Answers     []TestAttemptAnswer
	CompletedAt time.Time
}

type TestResult struct {
	AttemptID    int64
	Score        int
	Passed       bool
	PassingScore int
	CompletedAt  time.Time
	Answers      []TestAttemptAnswer
}

// TestSubmissionEffects are decisions made by the application, persisted together with an attempt.
type TestSubmissionEffects struct {
	AssessmentScope    string
	FormHash           string
	InstrumentSnapshot json.RawMessage
	AssessmentType     string
	CompletePage       bool
}
