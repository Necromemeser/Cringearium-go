package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestHandlerGetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	handler := NewHandler(application.NewCourses(repo))

	courses := []*domain.Course{{ID: 1, Title: "Math", Price: 99000, Status: domain.CourseStatusPublished}}
	repo.EXPECT().FindAll(gomock.Any()).Return(courses, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/courses", nil)
	rec := httptest.NewRecorder()
	handler.GetAll(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
}

func TestHandlerGetAllRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	handler := NewHandler(application.NewCourses(repo))

	repo.EXPECT().FindAll(gomock.Any()).Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/courses", nil)
	rec := httptest.NewRecorder()
	handler.GetAll(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHandlerGetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	handler := NewHandler(application.NewCourses(repo))

	course := &domain.CourseDetails{
		Course: domain.Course{ID: 1, Title: "Math", Status: domain.CourseStatusPublished},
		Sections: []domain.SectionDetails{{
			Section: domain.Section{ID: 10, Title: "Equations", Position: 0},
			Pages:   []domain.Page{{ID: 100, Title: "Theory", Type: domain.PageTypeTheory, Position: 0, Content: "content"}},
		}},
	}
	repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(course, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/courses/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	handler.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response courseDetailsResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != 1 || response.Title != "Math" {
		t.Fatalf("response = %#v", response)
	}
	if len(response.Sections) != 1 || len(response.Sections[0].Pages) != 1 {
		t.Fatalf("response sections/pages = %#v", response.Sections)
	}
}

func TestHandlerGetByIDInvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	handler := NewHandler(application.NewCourses(repo))

	req := httptest.NewRequest(http.MethodGet, "/api/courses/nope", nil)
	req.SetPathValue("id", "nope")
	rec := httptest.NewRecorder()
	handler.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandlerGetByIDNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	handler := NewHandler(application.NewCourses(repo))

	repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/courses/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	handler.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandlerEnroll(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	handler := NewHandler(application.NewCourses(repo))

	repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&domain.CourseDetails{Course: domain.Course{ID: 1, Status: domain.CourseStatusPublished}}, nil)
	repo.EXPECT().HasAccess(gomock.Any(), int64(42), int64(1)).Return(false, nil)
	repo.EXPECT().GrantAccess(gomock.Any(), int64(42), int64(1)).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/courses/1/enroll", nil)
	req.SetPathValue("id", "1")
	req.Header.Set("X-User-ID", "42")
	rec := httptest.NewRecorder()
	handler.Enroll(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestHandlerEnrollUnauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	handler := NewHandler(application.NewCourses(repo))

	req := httptest.NewRequest(http.MethodPost, "/api/courses/1/enroll", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	handler.Enroll(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandlerEnrollErrors(t *testing.T) {
	tests := []struct {
		name       string
		course     *domain.CourseDetails
		hasAccess  bool
		wantStatus int
	}{
		{
			name:       "not found",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "not published",
			course:     &domain.CourseDetails{Course: domain.Course{ID: 1, Status: domain.CourseStatusDraft}},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "paid",
			course:     &domain.CourseDetails{Course: domain.Course{ID: 1, Status: domain.CourseStatusPublished, Price: 99000}},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "already enrolled",
			course:     &domain.CourseDetails{Course: domain.Course{ID: 1, Status: domain.CourseStatusPublished}},
			hasAccess:  true,
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockCourseRepository(ctrl)
			handler := NewHandler(application.NewCourses(repo))

			repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(tt.course, nil)
			if tt.course != nil && tt.course.Course.Status == domain.CourseStatusPublished && tt.course.Course.Price == 0 {
				repo.EXPECT().HasAccess(gomock.Any(), int64(42), int64(1)).Return(tt.hasAccess, nil)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/courses/1/enroll", nil)
			req.SetPathValue("id", "1")
			req.Header.Set("X-User-ID", "42")
			rec := httptest.NewRecorder()
			handler.Enroll(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, strings.TrimSpace(rec.Body.String()))
			}
		})
	}
}

func TestHandlerGetAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	handler := NewHandler(application.NewCourses(repo))

	repo.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&domain.CourseDetails{Course: domain.Course{ID: 1}}, nil)
	repo.EXPECT().HasAccess(gomock.Any(), int64(42), int64(1)).Return(true, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/courses/1/access", nil)
	req.SetPathValue("id", "1")
	req.Header.Set("X-User-ID", "42")
	rec := httptest.NewRecorder()
	handler.GetAccess(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response map[string]bool
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response["has_access"] {
		t.Fatal("has_access = false, want true")
	}
}

func TestHandlerGetAccessUnauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockCourseRepository(ctrl)
	handler := NewHandler(application.NewCourses(repo))

	req := httptest.NewRequest(http.MethodGet, "/api/courses/1/access", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	handler.GetAccess(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestUserIDFromHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-ID", "42")

	got, err := userIDFromHeader(req)
	if err != nil {
		t.Fatalf("userIDFromHeader() error = %v", err)
	}
	if got != 42 {
		t.Fatalf("userIDFromHeader() = %d, want 42", got)
	}
}

func TestUserIDFromHeaderInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-ID", "abc")

	_, err := userIDFromHeader(req)
	if err == nil {
		t.Fatal("userIDFromHeader() error = nil, want error")
	}
}
