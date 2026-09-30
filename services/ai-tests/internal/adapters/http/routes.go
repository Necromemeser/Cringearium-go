package http

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("POST /api/adaptive-tests", handler.CreateSession)
	mux.HandleFunc("POST /api/adaptive-tests/{sessionId}/answers", handler.SubmitAnswers)
}
