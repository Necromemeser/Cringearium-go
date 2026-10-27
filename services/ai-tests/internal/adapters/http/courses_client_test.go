package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/ports"
)

func TestCoursesClientPreservesPracticeAndAnswerEvidence(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("topic_page_id") != "100" || r.Header.Get("X-User-ID") != "7" {
			t.Errorf("incorrect request %s", r.URL)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"course_id":42,"course_title":"Python","topic_title":"AI input practice","materials":[{"page_id":10,"title":"Input","content":"input returns string"}],"ordinary_tests":[{"id":5,"page_id":20,"title":"Input test","questions":[{"question":"input returns?","options":["string","integer"],"answered":true,"selected_answer":"integer","correct_answer":"string","is_correct":false}]}],"test_results":[{"test_id":5,"page_id":20,"title":"Input test","correct_count":1,"total_count":2}],"allowed_topics":[{"page_id":10,"title":"Input"}]}`)
	}))
	defer server.Close()
	client := NewCoursesClient(slog.New(slog.NewTextHandler(io.Discard, nil)), server.URL)
	scope := int64(100)
	result, err := client.GetCourseContext(context.Background(), 7, 42, &scope)
	if err != nil {
		t.Fatal(err)
	}
	if result.TopicTitle != "AI input practice" || len(result.TestResults) != 1 || result.TestResults[0].PageID != 20 || result.TestResults[0].Title != "Input test" {
		t.Fatalf("lost scope/result: %+v", result)
	}
	q := result.OrdinaryTests[0].Questions[0]
	if !q.Answered || q.IsCorrect || q.SelectedAnswer != "integer" || q.CorrectAnswer != "string" {
		t.Fatalf("lost answer: %+v", q)
	}
}

func TestCoursesClientRejectsInvalidScope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadRequest) }))
	defer server.Close()
	client := NewCoursesClient(slog.New(slog.NewTextHandler(io.Discard, nil)), server.URL)
	if _, err := client.GetCourseContext(context.Background(), 7, 42, nil); !errors.Is(err, ports.ErrInvalidCourseContext) {
		t.Fatal(err)
	}
}
