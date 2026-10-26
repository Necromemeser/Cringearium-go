package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/ports"
)

func (h *Handler) GetLearningProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	courseID, err := strconv.ParseInt(r.URL.Query().Get("course_id"), 10, 64)
	if err != nil || courseID <= 0 {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}
	var topic *int64
	if raw := r.URL.Query().Get("topic_page_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "invalid practice page", http.StatusBadRequest)
			return
		}
		topic = &id
	}
	profile, err := h.service.GetLearningProfile(r.Context(), userID, courseID, topic)
	if err != nil {
		if errors.Is(err, ports.ErrCourseUnavailable) {
			http.Error(w, "course is unavailable", http.StatusNotFound)
			return
		}
		if errors.Is(err, ports.ErrInvalidCourseContext) {
			http.Error(w, "invalid practice context", http.StatusBadRequest)
			return
		}
		h.logger.Error("load learning profile", "error", err)
		http.Error(w, "failed to load learning profile", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}
