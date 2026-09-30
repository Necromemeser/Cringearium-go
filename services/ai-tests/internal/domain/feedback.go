package domain

import "time"

type AdaptiveFeedback struct {
	SessionID string
	Summary   string

	MasteredTopics []TopicFeedback
	TopicsToReview []TopicFeedback
	NextSteps      []NextStep

	CreatedAt time.Time
}

type TopicFeedback struct {
	Title       string
	TopicPageID *int64
	Reason      string
}

type NextStep struct {
	Title       string
	Description string
	TopicPageID *int64
}
