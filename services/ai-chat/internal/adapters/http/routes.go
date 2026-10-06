package http

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("GET /api/chats", handler.ListConversations)
	mux.HandleFunc("POST /api/chats", handler.CreateConversation)
	mux.HandleFunc("GET /api/chats/{id}", handler.GetConversation)
	mux.HandleFunc("DELETE /api/chats/{id}", handler.DeleteConversation)
	mux.HandleFunc("GET /api/chats/{chatId}/messages", handler.GetMessages)
	mux.HandleFunc("GET /api/messages/chat/{chatId}", handler.GetMessages)
	mux.HandleFunc("POST /api/messages/send", handler.SendMessage)
	mux.HandleFunc("POST /api/deepseek", handler.DeepSeek)
	mux.HandleFunc("GET /internal/admin/chat-stats", handler.GetAdminStats)
}
