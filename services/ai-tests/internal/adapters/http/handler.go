package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
)

type Handler struct {
	logger  *slog.Logger
	service *application.AdaptiveTestService
}

func NewHandler(logger *slog.Logger, service *application.AdaptiveTestService) *Handler {
	return &Handler{logger: logger, service: service}
}

type createSessionRequest struct {
	CourseID      int64  `json:"course_id"`
	TopicPageID   *int64 `json:"topic_page_id,omitempty"`
	QuestionCount int    `json:"question_count"`
}

type submitAnswersRequest struct {
	RoundID int64            `json:"round_id"`
	Answers map[string]string `json:"answers"`
}

type sessionResponse struct {
	ID            string            `json:"id"`
	CourseID      int64             `json:"course_id"`
	TopicPageID   *int64            `json:"topic_page_id,omitempty"`
	Status        domain.SessionStatus `json:"status"`
	CurrentRound  int               `json:"current_round"`
	QuestionCount int               `json:"question_count"`
	Rounds        []roundResponse    `json:"rounds"`
	Feedback      *feedbackResponse  `json:"feedback,omitempty"`
}

type roundResponse struct {
	ID           int64                `json:"id"`
	RoundNumber  int                  `json:"round_number"`
	Strategy     domain.RoundStrategy `json:"strategy"`
	Status       domain.RoundStatus   `json:"status"`
	Title        string               `json:"title"`
	Instructions string               `json:"instructions,omitempty"`
	Questions    []questionResponse   `json:"questions"`
}

type questionResponse struct {
	ID             int64                  `json:"id"`
	Position       int                    `json:"position"`
	TopicPageID    *int64                 `json:"topic_page_id,omitempty"`
	TopicTitle     string                 `json:"topic_title,omitempty"`
	Question       string                 `json:"question"`
	Difficulty     int                    `json:"difficulty"`
	KnowledgeBasis domain.KnowledgeBasis `json:"knowledge_basis"`
	Sources        []domain.QuestionSource `json:"sources,omitempty"`
	Explanation    string                 `json:"explanation,omitempty"`
	Options        []optionResponse       `json:"options"`
}

type optionResponse struct {
	Key      string `json:"key"`
	Text     string `json:"text"`
	Position int    `json:"position"`
}

type feedbackResponse struct {
	Summary        string                 `json:"summary"`
	MasteredTopics []domain.TopicFeedback `json:"mastered_topics"`
	TopicsToReview []domain.TopicFeedback `json:"topics_to_review"`
	NextSteps      []domain.NextStep      `json:"next_steps"`
}

type adminAIStatsResponse struct {
	TotalSessions      int64            `json:"total_sessions"`
	CompletedSessions  int64            `json:"completed_sessions"`
	FailedSessions     int64            `json:"failed_sessions"`
	QuestionsGenerated int64            `json:"questions_generated"`
	AnswersSubmitted   int64            `json:"answers_submitted"`
	AccuracyPercent    float64          `json:"accuracy_percent"`
	RecentSessions     []adminAISession `json:"recent_sessions"`
}

type adminAISession struct {
	ID            string               `json:"id"`
	UserID        int64                `json:"user_id"`
	CourseID      int64                `json:"course_id"`
	TopicPageID   *int64               `json:"topic_page_id,omitempty"`
	Status        domain.SessionStatus `json:"status"`
	CurrentRound  int                  `json:"current_round"`
	QuestionCount int                  `json:"question_count"`
	CreatedAt     string               `json:"created_at"`
	CompletedAt   *string              `json:"completed_at,omitempty"`
	Summary       string               `json:"summary,omitempty"`
}

func (h *Handler) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetAdminStats(r.Context())
	if err != nil {
		h.logger.Error("failed to get adaptive admin stats", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := adminAIStatsResponse{
		TotalSessions: stats.TotalSessions,
		CompletedSessions: stats.CompletedSessions,
		FailedSessions: stats.FailedSessions,
		QuestionsGenerated: stats.QuestionsGenerated,
		AnswersSubmitted: stats.AnswersSubmitted,
		AccuracyPercent: stats.AccuracyPercent,
		RecentSessions: make([]adminAISession, 0, len(stats.RecentSessions)),
	}

	for _, session := range stats.RecentSessions {
		item := adminAISession{
			ID: session.ID,
			UserID: session.UserID,
			CourseID: session.CourseID,
			TopicPageID: session.TopicPageID,
			Status: session.Status,
			CurrentRound: session.CurrentRound,
			QuestionCount: session.QuestionCount,
			CreatedAt: session.CreatedAt.Format(time.RFC3339),
			Summary: session.Summary,
		}
		if session.CompletedAt != nil {
			value := session.CompletedAt.Format(time.RFC3339)
			item.CompletedAt = &value
		}
		response.RecentSessions = append(response.RecentSessions, item)
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		h.logger.Warn("create adaptive session unauthorized", "error", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var request createSessionRequest
	if err := decodeJSON(r, &request); err != nil {
		h.logger.Warn("create adaptive session invalid body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("creating adaptive session",
		"user_id", userID,
		"course_id", request.CourseID,
		"topic_page_id", request.TopicPageID,
		"question_count", request.QuestionCount,
	)

	session, err := h.service.CreateSession(
		r.Context(),
		userID,
		request.CourseID,
		request.TopicPageID,
		request.QuestionCount,
	)
	if err != nil {
		h.logger.Error("failed to create adaptive session",
			"user_id", userID,
			"course_id", request.CourseID,
			"error", err,
		)
		switch {
		case errors.Is(err, application.ErrInvalidQuestionCount):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, application.ErrInvalidGeneratedRound):
			http.Error(w, "generated test failed validation", http.StatusBadGateway)
		default:
			http.Error(w, "failed to create adaptive test", http.StatusInternalServerError)
		}
		return
	}

	h.logger.Info("adaptive session created", "session_id", session.ID, "user_id", userID)
	writeJSON(w, http.StatusCreated, toSessionResponse(session))
}

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		h.logger.Warn("get adaptive session unauthorized", "error", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := r.PathValue("sessionId")
	if sessionID == "" {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	session, err := h.service.GetSession(r.Context(), userID, sessionID)
	if err != nil {
		h.logger.Error("failed to get adaptive session",
			"session_id", sessionID,
			"user_id", userID,
			"error", err,
		)
		if errors.Is(err, application.ErrRoundNotFound) {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get adaptive test", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toSessionResponse(session))
}

func (h *Handler) SubmitAnswers(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		h.logger.Warn("submit adaptive answers unauthorized", "error", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := r.PathValue("sessionId")
	if sessionID == "" {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	var request submitAnswersRequest
	if err := decodeJSON(r, &request); err != nil {
		h.logger.Warn("submit adaptive answers invalid body", "session_id", sessionID, "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	answers := make(map[int64]string, len(request.Answers))
	for questionID, optionKey := range request.Answers {
		id, err := strconv.ParseInt(questionID, 10, 64)
		if err != nil || id <= 0 || optionKey == "" {
			h.logger.Warn("submit adaptive answers invalid answer",
				"session_id", sessionID,
				"question_id", questionID,
			)
			http.Error(w, "invalid answers", http.StatusBadRequest)
			return
		}
		answers[id] = optionKey
	}

	h.logger.Info("submitting adaptive answers",
		"session_id", sessionID,
		"user_id", userID,
		"round_id", request.RoundID,
		"answer_count", len(answers),
	)

	session, err := h.service.SubmitAnswers(
		r.Context(),
		userID,
		sessionID,
		request.RoundID,
		answers,
	)
	if err != nil {
		h.logger.Error("failed to submit adaptive answers",
			"session_id", sessionID,
			"user_id", userID,
			"round_id", request.RoundID,
			"error", err,
		)
		switch {
		case errors.Is(err, application.ErrSessionNotInProgress):
			http.Error(w, "session is not in progress", http.StatusConflict)
		case errors.Is(err, application.ErrRoundNotFound):
			http.Error(w, "round not found", http.StatusNotFound)
		case errors.Is(err, application.ErrInvalidAnswer):
			http.Error(w, "invalid answers", http.StatusBadRequest)
		default:
			http.Error(w, "failed to submit answers", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, toSessionResponse(session))
}

func toSessionResponse(session domain.AdaptiveSession) sessionResponse {
	response := sessionResponse{
		ID:            session.ID,
		CourseID:      session.CourseID,
		TopicPageID:   session.TopicPageID,
		Status:        session.Status,
		CurrentRound:  session.CurrentRound,
		QuestionCount: session.QuestionCount,
		Rounds:        make([]roundResponse, 0, len(session.Rounds)),
	}

	for _, round := range session.Rounds {
		rr := roundResponse{
			ID:           round.ID,
			RoundNumber:  round.RoundNumber,
			Strategy:     round.Strategy,
			Status:       round.Status,
			Title:        round.Title,
			Instructions: round.Instructions,
			Questions:    make([]questionResponse, 0, len(round.Questions)),
		}

		for _, question := range round.Questions {
			qr := questionResponse{
				ID:             question.ID,
				Position:       question.Position,
				TopicPageID:    question.TopicPageID,
				TopicTitle:     question.TopicTitle,
				Question:       question.Question,
				Difficulty:     question.Difficulty,
				KnowledgeBasis: question.KnowledgeBasis,
				Sources:        question.Sources,
				Explanation:    explanationForRound(round.Status, question.Explanation),
				Options:        make([]optionResponse, 0, len(question.Options)),
			}

			for _, option := range question.Options {
				qr.Options = append(qr.Options, optionResponse{
					Key:      option.Key,
					Text:     option.Text,
					Position: option.Position,
				})
			}

			rr.Questions = append(rr.Questions, qr)
		}

		response.Rounds = append(response.Rounds, rr)
	}

	if session.Feedback != nil {
		response.Feedback = &feedbackResponse{
			Summary:        session.Feedback.Summary,
			MasteredTopics: session.Feedback.MasteredTopics,
			TopicsToReview: session.Feedback.TopicsToReview,
			NextSteps:      session.Feedback.NextSteps,
		}
	}

	return response
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func userIDFromHeader(r *http.Request) (int64, error) {
	value := r.Header.Get("X-User-ID")
	if value == "" {
		return 0, errors.New("missing user id")
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid user id")
	}

	return id, nil
}

func explanationForRound(status domain.RoundStatus, explanation string) string {
	if status != domain.RoundCompleted {
		return ""
	}
	return explanation
}
