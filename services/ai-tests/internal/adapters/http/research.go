package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/domain"
)

type ResearchEventsReader interface {
	ResearchEvents(context.Context, []int64) ([]domain.ResearchEvent, bool, error)
}

func ResearchEventsHandler(reader ResearchEventsReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ids []int64
		for _, value := range strings.Split(r.URL.Query().Get("courses"), ",") {
			id, err := strconv.ParseInt(value, 10, 64)
			if err != nil || id <= 0 {
				http.Error(w, "invalid courses", 400)
				return
			}
			ids = append(ids, id)
		}
		if len(ids) > 2 || len(ids) == 0 {
			http.Error(w, "invalid course count", 400)
			return
		}
		events, truncated, err := reader.ResearchEvents(r.Context(), ids)
		if err != nil {
			http.Error(w, "failed to load research events", 500)
			return
		}
		writeJSON(w, 200, struct {
			Events    []domain.ResearchEvent `json:"events"`
			Truncated bool                   `json:"truncated"`
		}{events, truncated})
	}
}
