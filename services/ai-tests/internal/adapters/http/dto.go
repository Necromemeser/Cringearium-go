package http

type CreateAdaptiveSessionRequest struct {
	CourseID      int64  `json:"course_id"`
	TopicPageID   *int64 `json:"topic_page_id,omitempty"`
	QuestionCount int    `json:"question_count"`
}

type AdaptiveSessionResponse struct {
	ID           string               `json:"id"`
	CourseID     int64                `json:"course_id"`
	TopicPageID  *int64               `json:"topic_page_id,omitempty"`
	Status       string               `json:"status"`
	CurrentRound int                  `json:"current_round"`
	Round        *AdaptiveRoundDTO    `json:"round,omitempty"`
	Feedback     *AdaptiveFeedbackDTO `json:"feedback,omitempty"`
	CreatedAt    string               `json:"created_at"`
}

type AdaptiveRoundDTO struct {
	ID           int64                 `json:"id"`
	RoundNumber  int                   `json:"round_number"`
	Strategy     string                `json:"strategy"`
	Title        string                `json:"title"`
	Instructions string                `json:"instructions,omitempty"`
	Questions    []AdaptiveQuestionDTO `json:"questions"`
}

type AdaptiveQuestionDTO struct {
	ID             int64               `json:"id"`
	Position       int                 `json:"position"`
	TopicTitle     string              `json:"topic_title,omitempty"`
	Question       string              `json:"question"`
	Difficulty     int                 `json:"difficulty"`
	KnowledgeBasis string              `json:"knowledge_basis"`
	Options        []AdaptiveOptionDTO `json:"options"`
}

type AdaptiveOptionDTO struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

type SubmitAdaptiveAnswersRequest struct {
	Answers []SubmittedAdaptiveAnswer `json:"answers"`
}

type SubmittedAdaptiveAnswer struct {
	QuestionID int64  `json:"question_id"`
	OptionKey  string `json:"option_key"`
}

type AdaptiveRoundResultDTO struct {
	RoundNumber  int                       `json:"round_number"`
	CorrectCount int                       `json:"correct_count"`
	TotalCount   int                       `json:"total_count"`
	ScorePercent int                       `json:"score_percent"`
	Answers      []AdaptiveAnswerResultDTO `json:"answers"`
}

type AdaptiveAnswerResultDTO struct {
	QuestionID        int64  `json:"question_id"`
	SelectedOptionKey string `json:"selected_option_key"`
	CorrectOptionKey  string `json:"correct_option_key"`
	IsCorrect         bool   `json:"is_correct"`
	Explanation       string `json:"explanation"`
}

type AdaptiveFeedbackDTO struct {
	Summary        string             `json:"summary"`
	MasteredTopics []TopicFeedbackDTO `json:"mastered_topics"`
	TopicsToReview []TopicFeedbackDTO `json:"topics_to_review"`
	NextSteps      []NextStepDTO      `json:"next_steps"`
}

type TopicFeedbackDTO struct {
	Title       string `json:"title"`
	TopicPageID *int64 `json:"topic_page_id,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

type NextStepDTO struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	TopicPageID *int64 `json:"topic_page_id,omitempty"`
}
