package domain

import "time"

type ResearchEvent struct {
	SessionID      string     `json:"session_id" db:"session_id"`
	UserID         int64      `json:"user_id" db:"user_id"`
	CourseID       int64      `json:"course_id" db:"course_id"`
	SessionStatus  string     `json:"session_status" db:"session_status"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	CompletedAt    *time.Time `json:"completed_at" db:"completed_at"`
	RoundID        *int64     `json:"round_id" db:"round_id"`
	RoundNumber    *int       `json:"round_number" db:"round_number"`
	QuestionID     *int64     `json:"question_id" db:"question_id"`
	QuestionText   *string    `json:"question_text" db:"question_text"`
	CorrectOption  *string    `json:"correct_option" db:"correct_option"`
	Reason         *string    `json:"reason" db:"reason"`
	SkillCode      *string    `json:"skill_code" db:"skill_code"`
	Difficulty     *int       `json:"difficulty" db:"difficulty"`
	SelectedOption *string    `json:"selected_option" db:"selected_option"`
	IsCorrect      *bool      `json:"is_correct" db:"is_correct"`
	AnsweredAt     *time.Time `json:"answered_at" db:"answered_at"`
	PromptVersion  *string    `json:"prompt_version" db:"prompt_version"`
	Model          *string    `json:"model" db:"model"`
	PolicyVersion  *string    `json:"policy_version" db:"policy_version"`
	Contributes    *bool      `json:"contributes" db:"contributes"`
}
