package domain

import (
	"encoding/json"
	"time"
)

type AdaptiveFeedback struct {
	SessionID string
	Summary   string

	MasteredTopics []TopicFeedback
	TopicsToReview []TopicFeedback
	NextSteps      []NextStep

	CreatedAt time.Time
}

type TopicFeedback struct {
	Title       string `json:"title"`
	TopicPageID *int64 `json:"topic_page_id,omitempty"`
	Reason      string `json:"reason"`
}

type NextStep struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	TopicPageID *int64 `json:"topic_page_id,omitempty"`
}

func (t *TopicFeedback) UnmarshalJSON(data []byte) error {
	var wire struct {
		Title       string
		Reason      string
		TopicPageID *int64
		SnakeID     *int64 `json:"topic_page_id"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	t.Title, t.Reason, t.TopicPageID = wire.Title, wire.Reason, wire.TopicPageID
	if wire.SnakeID != nil {
		t.TopicPageID = wire.SnakeID
	}
	return nil
}
func (n *NextStep) UnmarshalJSON(data []byte) error {
	var wire struct {
		Title       string
		Description string
		TopicPageID *int64
		SnakeID     *int64 `json:"topic_page_id"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	n.Title, n.Description, n.TopicPageID = wire.Title, wire.Description, wire.TopicPageID
	if wire.SnakeID != nil {
		n.TopicPageID = wire.SnakeID
	}
	return nil
}
