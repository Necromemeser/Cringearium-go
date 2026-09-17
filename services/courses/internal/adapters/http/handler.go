package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Necromemeser/Cringearium-go/services/courses/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/domain"
)

type Handler struct { courses *application.Courses }
func NewHandler(courses *application.Courses) *Handler { return &Handler{courses: courses} }

type courseResponse struct { ID int64 `json:"id"`; Title string `json:"title"`; Theme string `json:"theme,omitempty"`; Description string `json:"description,omitempty"`; Price int `json:"price"`; ImageID string `json:"image_id,omitempty"`; AuthorID *int64 `json:"author_id,omitempty"`; Status domain.CourseStatus `json:"status"` }
type courseDetailsResponse struct { ID int64 `json:"id"`; Title string `json:"title"`; Theme string `json:"theme,omitempty"`; Description string `json:"description,omitempty"`; Price int `json:"price"`; ImageID string `json:"image_id,omitempty"`; AuthorID *int64 `json:"author_id,omitempty"`; Status domain.CourseStatus `json:"status"`; Sections []sectionDetailsResponse `json:"sections"` }
type sectionDetailsResponse struct { ID int64 `json:"id"`; Title string `json:"title"`; Description string `json:"description,omitempty"`; Position int `json:"position"`; Pages []pageResponse `json:"pages"` }
type pageResponse struct { ID int64 `json:"id"`; Title string `json:"title"`; Type domain.PageType `json:"type"`; Content string `json:"content,omitempty"`; Position int `json:"position"` }

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	courses, err := h.courses.GetAll(r.Context()); if err != nil { http.Error(w, "internal server error", http.StatusInternalServerError); return }
	response := make([]courseResponse, 0, len(courses)); for _, course := range courses { response = append(response, toCourseResponse(course)) }; writeJSON(w, http.StatusOK, response)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64); if err != nil || id <= 0 { http.Error(w, "invalid course id", http.StatusBadRequest); return }
	course, err := h.courses.GetByID(r.Context(), id); if err != nil { if errors.Is(err, application.ErrCourseNotFound) { http.Error(w, "course not found", http.StatusNotFound); return }; http.Error(w, "internal server error", http.StatusInternalServerError); return }
	response := courseDetailsResponse{ID: course.Course.ID, Title: course.Course.Title, Theme: course.Course.Theme, Description: course.Course.Description, Price: course.Course.Price, AuthorID: course.Course.AuthorID, Status: course.Course.Status, Sections: make([]sectionDetailsResponse, 0, len(course.Sections))}
	if course.Course.ImageID != nil { response.ImageID = *course.Course.ImageID }
	for _, section := range course.Sections { sr := sectionDetailsResponse{ID: section.Section.ID, Title: section.Section.Title, Description: section.Section.Description, Position: section.Section.Position, Pages: make([]pageResponse, 0, len(section.Pages))}; for _, page := range section.Pages { sr.Pages = append(sr.Pages, pageResponse{ID: page.ID, Title: page.Title, Type: page.Type, Content: page.Content, Position: page.Position}) }; response.Sections = append(response.Sections, sr) }
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) Enroll(w http.ResponseWriter, r *http.Request) {
	courseID, err := strconv.ParseInt(r.PathValue("id"), 10, 64); if err != nil || courseID <= 0 { http.Error(w, "invalid course id", http.StatusBadRequest); return }
	userID, err := userIDFromHeader(r); if err != nil { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
	if err = h.courses.Enroll(r.Context(), userID, courseID); err != nil { switch { case errors.Is(err, application.ErrCourseNotFound): http.Error(w, "course not found", http.StatusNotFound); case errors.Is(err, application.ErrCourseNotPublished): http.Error(w, "course is not published", http.StatusConflict); case errors.Is(err, application.ErrCourseNotFree): http.Error(w, "paid courses must be purchased", http.StatusConflict); case errors.Is(err, application.ErrAlreadyHasAccess): http.Error(w, "already enrolled", http.StatusConflict); default: http.Error(w, "internal server error", http.StatusInternalServerError) }; return }
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetAccess(w http.ResponseWriter, r *http.Request) {
	courseID, err := strconv.ParseInt(r.PathValue("id"), 10, 64); if err != nil || courseID <= 0 { http.Error(w, "invalid course id", http.StatusBadRequest); return }
	userID, err := userIDFromHeader(r); if err != nil { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
	hasAccess, err := h.courses.HasAccess(r.Context(), userID, courseID); if err != nil { if errors.Is(err, application.ErrCourseNotFound) { http.Error(w, "course not found", http.StatusNotFound); return }; http.Error(w, "internal server error", http.StatusInternalServerError); return }
	writeJSON(w, http.StatusOK, map[string]bool{"has_access": hasAccess})
}

func (h *Handler) GetEnrolled(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r); if err != nil { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
	courses, err := h.courses.FindEnrolled(r.Context(), userID); if err != nil { http.Error(w, "internal server error", http.StatusInternalServerError); return }
	response := make([]courseResponse, 0, len(courses)); for _, course := range courses { response = append(response, toCourseResponse(course)) }; writeJSON(w, http.StatusOK, response)
}

func (h *Handler) GetProgress(w http.ResponseWriter, r *http.Request) {
	courseID, err := strconv.ParseInt(r.PathValue("id"), 10, 64); if err != nil || courseID <= 0 { http.Error(w, "invalid course id", http.StatusBadRequest); return }
	userID, err := userIDFromHeader(r); if err != nil { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
	hasAccess, err := h.courses.HasAccess(r.Context(), userID, courseID); if err != nil { if errors.Is(err, application.ErrCourseNotFound) { http.Error(w, "course not found", http.StatusNotFound); return }; http.Error(w, "internal server error", http.StatusInternalServerError); return }
	if !hasAccess { http.Error(w, "course access required", http.StatusForbidden); return }
	pageIDs, err := h.courses.GetCompletedPages(r.Context(), userID, courseID); if err != nil { http.Error(w, "internal server error", http.StatusInternalServerError); return }
	writeJSON(w, http.StatusOK, map[string][]int64{"completed_page_ids": pageIDs})
}

func (h *Handler) CompletePage(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.ParseInt(r.PathValue("pageId"), 10, 64); if err != nil || pageID <= 0 { http.Error(w, "invalid page id", http.StatusBadRequest); return }
	userID, err := userIDFromHeader(r); if err != nil { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
	if err = h.courses.CompletePage(r.Context(), userID, pageID); err != nil { http.Error(w, "internal server error", http.StatusInternalServerError); return }
	w.WriteHeader(http.StatusNoContent)
}

func userIDFromHeader(r *http.Request) (int64, error) { return strconv.ParseInt(r.Header.Get("X-User-ID"), 10, 64) }
func toCourseResponse(course *domain.Course) courseResponse { response := courseResponse{ID: course.ID, Title: course.Title, Theme: course.Theme, Description: course.Description, Price: course.Price, AuthorID: course.AuthorID, Status: course.Status}; if course.ImageID != nil { response.ImageID = *course.ImageID }; return response }
func writeJSON(w http.ResponseWriter, status int, value any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(value) }
