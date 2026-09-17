package domain

import "time"

type Test struct {
	ID           int64
	PageID       int64
	PassingScore int
	Questions    []TestQuestion
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
	QuestionID     int64
	AnswerID       int64
	CorrectAnswerID int64
	IsCorrect      bool
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
