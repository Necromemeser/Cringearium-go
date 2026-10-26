package domain

import (
	"encoding/json"
	"time"
)

type AssessmentResult struct {
	FormHash           *string         `json:"form_hash"`
	InstrumentSnapshot json.RawMessage `json:"instrument_snapshot,omitempty"`
	ID                 int64           `json:"id"`
	UserID             int64           `json:"user_id"`
	CourseID           int64           `json:"course_id"`
	CourseTitle        string          `json:"course_title"`
	Type               string          `json:"type"`
	Scope              string          `json:"scope"`
	Score              int             `json:"score"`
	Answers            json.RawMessage `json:"answers"`
	CompletedAt        time.Time       `json:"completed_at"`
}
