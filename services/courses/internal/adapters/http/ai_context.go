package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

type aiContextResponse struct {
	Skills        []domain.LearningSkill `json:"skills"`
	CourseID      int64                  `json:"course_id"`
	TopicTitle    string                 `json:"topic_title"`
	CourseTitle   string                 `json:"course_title"`
	Materials     []aiContextMaterial    `json:"materials"`
	OrdinaryTests []aiContextTest        `json:"ordinary_tests"`
	TestResults   []aiContextResult      `json:"test_results"`
	AllowedTopics []aiContextTopic       `json:"allowed_topics"`
}

type aiContextMaterial struct {
	PageID  int64  `json:"page_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type aiContextTest struct {
	ID        int64               `json:"id"`
	PageID    int64               `json:"page_id"`
	Title     string              `json:"title"`
	Questions []aiContextQuestion `json:"questions"`
}

type aiContextQuestion struct {
	ID             int64    `json:"id"`
	SelectedAnswer string   `json:"selected_answer"`
	CorrectAnswer  string   `json:"correct_answer"`
	Answered       bool     `json:"answered"`
	IsCorrect      bool     `json:"is_correct"`
	Question       string   `json:"question"`
	Options        []string `json:"options"`
}

type aiContextResult struct {
	TestID       int64  `json:"test_id"`
	PageID       int64  `json:"page_id"`
	Title        string `json:"title"`
	CorrectCount int    `json:"correct_count"`
	TotalCount   int    `json:"total_count"`
}

type aiContextTopic struct {
	PageID int64  `json:"page_id"`
	Title  string `json:"title"`
}

func (h *Handler) GetAIContext(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	courseID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || courseID <= 0 {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	var topicPageID *int64
	if value := r.URL.Query().Get("topic_page_id"); value != "" {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "invalid topic page id", http.StatusBadRequest)
			return
		}
		topicPageID = &id
	}

	context, err := h.courses.GetAIContext(r.Context(), userID, courseID, topicPageID)
	if err != nil {
		if errors.Is(err, application.ErrPageNotInCourse) {
			http.Error(w, "invalid practice page", http.StatusBadRequest)
			return
		}
		if errors.Is(err, application.ErrCourseNotFound) {
			http.Error(w, "course not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toAIContextResponse(context))
}

func toAIContextResponse(context *domain.AIContext) aiContextResponse {
	response := aiContextResponse{
		Skills:        context.Skills,
		CourseID:      context.CourseID,
		CourseTitle:   context.CourseTitle,
		TopicTitle:    context.TopicTitle,
		Materials:     make([]aiContextMaterial, 0, len(context.Materials)),
		OrdinaryTests: make([]aiContextTest, 0, len(context.OrdinaryTests)),
		TestResults:   make([]aiContextResult, 0, len(context.TestResults)),
		AllowedTopics: make([]aiContextTopic, 0, len(context.AllowedTopics)),
	}

	for _, material := range context.Materials {
		response.Materials = append(response.Materials, aiContextMaterial{
			PageID:  material.PageID,
			Title:   material.Title,
			Content: material.Content,
		})
	}

	for _, test := range context.OrdinaryTests {
		item := aiContextTest{
			ID:        test.ID,
			PageID:    test.PageID,
			Title:     test.Title,
			Questions: make([]aiContextQuestion, 0, len(test.Questions)),
		}
		for _, question := range test.Questions {
			item.Questions = append(item.Questions, aiContextQuestion{
				ID: question.ID, SelectedAnswer: question.SelectedAnswer, CorrectAnswer: question.CorrectAnswer, Answered: question.Answered, IsCorrect: question.IsCorrect,
				Question: question.Question,
				Options:  question.Options,
			})
		}
		response.OrdinaryTests = append(response.OrdinaryTests, item)
	}

	for _, result := range context.TestResults {
		response.TestResults = append(response.TestResults, aiContextResult{
			TestID: result.TestID,
			PageID: result.PageID, Title: result.Title,
			CorrectCount: result.CorrectCount,
			TotalCount:   result.TotalCount,
		})
	}

	for _, topic := range context.AllowedTopics {
		response.AllowedTopics = append(response.AllowedTopics, aiContextTopic{
			PageID: topic.PageID,
			Title:  topic.Title,
		})
	}

	return response
}
