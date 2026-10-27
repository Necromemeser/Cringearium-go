package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/ports"
)

type profileUseCases struct {
	ports.AdaptiveTestServiceUseCases
	calls int
}

func (p *profileUseCases) GetLearningProfile(_ context.Context, userID, courseID int64, topic *int64) (domain.LearningProfile, error) {
	p.calls++
	return domain.BuildProfile([]domain.LearningSkill{{Code: "math.linear.balance", Title: "Linear", PageIDs: []int64{10}}}, nil, 5, time.Now()), nil
}
func TestProfileRouteIsAuthorizedAndNotTreatedAsSessionID(t *testing.T) {
	service := &profileUseCases{}
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), service)
	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)
	for _, tc := range []struct {
		user, path string
		status     int
	}{{"", "/api/adaptive-tests/profile?course_id=42", 401}, {"7", "/api/adaptive-tests/profile?course_id=bad", 400}, {"7", "/api/adaptive-tests/profile?course_id=42&topic_page_id=100", 200}} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		if tc.user != "" {
			req.Header.Set("X-User-ID", tc.user)
		}
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, req)
		if response.Code != tc.status {
			t.Fatalf("status %d want %d: %s", response.Code, tc.status, response.Body.String())
		}
		if tc.status == 200 && (!strings.Contains(response.Body.String(), `"insufficient_data"`) || !strings.Contains(response.Body.String(), `"diagnose"`)) {
			t.Fatal(response.Body.String())
		}
	}
	if service.calls != 1 {
		t.Fatalf("invalid requests reached use case: %d", service.calls)
	}
}

func TestCompletedResponseHasReadableRussianFeedbackAndRoundResults(t *testing.T) {
	session := domain.AdaptiveSession{Status: domain.SessionCompleted, Feedback: &domain.AdaptiveFeedback{Summary: "Round 1 was strong"}, Rounds: []domain.AdaptiveRound{{ID: 1, Status: domain.RoundCompleted, LearningPlan: []domain.QuestionTarget{{Reason: "prerequisite"}}, Questions: []domain.AdaptiveQuestion{{ID: 1, TopicTitle: "Линейные уравнения"}}, Answers: []domain.AdaptiveAnswer{{QuestionID: 1, IsCorrect: true}}}}}
	response := toSessionResponse(session)
	if response.Rounds[0].Purpose != "prerequisite_review" || response.Rounds[0].CorrectCount != 1 || response.Rounds[0].AnsweredCount != 1 {
		t.Fatalf("wrong round result: %+v", response.Rounds)
	}
	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	topic := payload["feedback"].(map[string]any)["mastered_topics"].([]any)[0].(map[string]any)
	if topic["title"] != "Линейные уравнения" || topic["reason"] == nil {
		t.Fatalf("empty result card: %s", raw)
	}
	if strings.Contains(string(raw), "Round 1") {
		t.Fatalf("untranslated summary: %s", raw)
	}
}
