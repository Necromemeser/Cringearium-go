package http

import (
	"context"
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
