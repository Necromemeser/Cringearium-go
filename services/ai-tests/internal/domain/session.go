package domain

import "time"

type SessionStatus string

const (
	SessionInProgress SessionStatus = "in_progress"
	SessionCompleted  SessionStatus = "completed"
	SessionFailed     SessionStatus = "failed"
)

type AdaptiveSession struct {
	ID              string
	UserID          int64
	CourseID        int64
	TopicPageID     *int64
	Status          SessionStatus
	CurrentRound    int
	QuestionCount   int
	ContextSnapshot []byte
	CreatedAt       time.Time
	CompletedAt     *time.Time

	Rounds   []AdaptiveRound
	Feedback *AdaptiveFeedback
}
