package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

type testResponse struct {
	ID           int64            `json:"id"`
	PageID       int64            `json:"page_id"`
	PassingScore int              `json:"passing_score"`
	Questions    []questionResponse `json:"questions"`
}

type questionResponse struct {
	ID       int64          `json:"id"`
	Question string         `json:"question"`
	Position int            `json:"position"`
	Answers  []answerResponse `json:"answers"`
}

type answerResponse struct {
	ID       int64  `json:"id"`
	Text     string `json:"text"`
	Position int    `json:"position"`
}

type submitTestRequest struct {
	Answers []submittedAnswer `json:"answers"`
}

type submittedAnswer struct {
	QuestionID int64 `json:"question_id"`
	AnswerID   int64 `json:"answer_id"`
}

type testResultResponse struct {
	AttemptID    int64 `json:"attempt_id"`
	Score        int   `json:"score"`
	Passed       bool  `json:"passed"`
	PassingScore int   `json:"passing_score"`
}

func (h *Handler) GetTest(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.ParseInt(r.PathValue("pageId"), 10, 64)
	if err != nil || pageID <= 0 {
		http.Error(w, "invalid page id", http.StatusBadRequest)
		return
	}
	userID, err := userIDFromHeader(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	test, err := h.courses.GetTest(r.Context(), userID, pageID)
	if err != nil {
		if errors.Is(err, application.ErrTestNotFound) {
			http.Error(w, "test not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := testResponse{
		ID:           test.ID,
		PageID:       test.PageID,
		PassingScore: test.PassingScore,
		Questions:    make([]questionResponse, 0, len(test.Questions)),
	}
	for _, question := range test.Questions {
		qr := questionResponse{
			ID:       question.ID,
			Question: question.Question,
			Position: question.Position,
			Answers:  make([]answerResponse, 0, len(question.Answers)),
		}
		for _, answer := range question.Answers {
			qr.Answers = append(qr.Answers, answerResponse{ID: answer.ID, Text: answer.Text, Position: answer.Position})
		}
		response.Questions = append(response.Questions, qr)
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) SubmitTest(w http.ResponseWriter, r *http.Request) {
	testID, err := strconv.ParseInt(r.PathValue("testId"), 10, 64)
	if err != nil || testID <= 0 {
		http.Error(w, "invalid test id", http.StatusBadRequest)
		return
	}
	userID, err := userIDFromHeader(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var request submitTestRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	answers := make([]domain.TestAttemptAnswer, 0, len(request.Answers))
	for _, answer := range request.Answers {
		answers = append(answers, domain.TestAttemptAnswer{QuestionID: answer.QuestionID, AnswerID: answer.AnswerID})
	}

	result, err := h.courses.SubmitTest(r.Context(), userID, testID, answers)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrTestNotFound):
			http.Error(w, "test not found", http.StatusNotFound)
		case errors.Is(err, application.ErrInvalidTestAnswers):
			http.Error(w, "invalid test answers", http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, testResultResponse{AttemptID: result.AttemptID, Score: result.Score, Passed: result.Passed, PassingScore: result.PassingScore})
}
