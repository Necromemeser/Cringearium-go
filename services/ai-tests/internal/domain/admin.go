package domain

import "time"

type AdminAIStats struct {
	TotalSessions       int64
	CompletedSessions   int64
	FailedSessions      int64
	QuestionsGenerated  int64
	AnswersSubmitted    int64
	AccuracyPercent     float64
	RecentSessions      []AdminAISession
}

type AdminAISession struct {
	ID              string
	UserID          int64
	CourseID        int64
	TopicPageID     *int64
	Status          SessionStatus
	CurrentRound    int
	QuestionCount   int
	CreatedAt       time.Time
	CompletedAt     *time.Time
	Summary         string
}
