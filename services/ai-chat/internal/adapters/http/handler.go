package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/domain"
)

type Handler struct {
	service *application.ChatService
}

func NewHandler(service *application.ChatService) *Handler {
	return &Handler{service: service}
}

type conversationRequest struct {
	Name string `json:"name"`
}

type messageRequest struct {
	Content string `json:"content"`
}

type conversationResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"chatName"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type messageResponse struct {
	ID           int64  `json:"id"`
	Content      string `json:"content"`
	IsAIResponse bool   `json:"isAiResponse"`
	Timestamp    string `json:"timestamp"`
	UserID       *int64 `json:"userId,omitempty"`
}

type adminStatsResponse struct {
	Conversations int64 `json:"conversations"`
	Messages      int64 `json:"messages"`
	UserMessages  int64 `json:"user_messages"`
	AIMessages    int64 `json:"ai_messages"`
	ActiveUsers   int64 `json:"active_users"`
}

type streamEvent struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Message string `json:"message,omitempty"`
}

func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req conversationRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
	}

	item, err := h.service.CreateConversation(r.Context(), userID, req.Name)
	if err != nil {
		slog.Error("create conversation failed", "error", err, "user_id", userID)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, toConversationResponse(item))
}

func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	items, err := h.service.ListConversations(r.Context(), userID)
	if err != nil {
		slog.Error("list conversations failed", "error", err, "user_id", userID)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := make([]conversationResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toConversationResponse(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	item, err := h.service.GetConversation(r.Context(), id, userID)
	if errors.Is(err, application.ErrConversationNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("get conversation failed", "error", err, "user_id", userID, "chat_id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, toConversationResponse(item))
}

func (h *Handler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteConversation(r.Context(), id, userID); err != nil {
		if errors.Is(err, application.ErrConversationNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	items, err := h.service.ListMessages(r.Context(), id, userID)
	if errors.Is(err, application.ErrConversationNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("list messages failed", "error", err, "user_id", userID, "chat_id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := make([]messageResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toMessageResponse(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	chatID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	var req messageRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32*1024))
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	input := strings.TrimSpace(req.Content)
	if input == "" {
		http.Error(w, "input is empty", http.StatusBadRequest)
		return
	}

	if _, err := h.service.GetConversation(r.Context(), chatID, userID); err != nil {
		if errors.Is(err, application.ErrConversationNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	err = h.service.StreamResponse(r.Context(), chatID, userID, input, func(chunk string) error {
		return writeSSE(w, flusher, streamEvent{Type: "chunk", Content: chunk})
	})
	if err != nil {
		slog.Error("stream response failed", "error", err, "user_id", userID, "chat_id", chatID)
		_ = writeSSE(w, flusher, streamEvent{Type: "error", Message: "не удалось получить ответ от ассистента"})
		return
	}

	_ = writeSSE(w, flusher, streamEvent{Type: "done"})
}

func (h *Handler) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetAdminStats(r.Context())
	if err != nil {
		slog.Error("get admin chat stats failed", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, adminStatsResponse{
		Conversations: stats.Conversations,
		Messages:      stats.Messages,
		UserMessages:  stats.UserMessages,
		AIMessages:    stats.AIMessages,
		ActiveUsers:   stats.ActiveUsers,
	})
}

func userID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.Header.Get("X-User-ID"), 10, 64)
	return id, err == nil && id > 0
}

func toConversationResponse(item domain.Conversation) conversationResponse {
	return conversationResponse{
		ID:        item.ID,
		Name:      item.Name,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.Format(time.RFC3339),
	}
}

func toMessageResponse(item domain.Message) messageResponse {
	return messageResponse{
		ID:           item.ID,
		Content:      item.Content,
		IsAIResponse: item.IsAIResponse,
		Timestamp:    item.CreatedAt.Format(time.RFC3339),
		UserID:       item.UserID,
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event streamEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte("data: " + string(data) + "\n\n")); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
