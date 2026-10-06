package http

import (
	"encoding/json"
	"io"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/domain"
)

type Handler struct{ service *application.ChatService }

func NewHandler(service *application.ChatService) *Handler { return &Handler{service: service} }

type conversationRequest struct {
	Name string `json:"name"`
	ChatName string `json:"chatName"`
}
type conversationResponse struct {
	ID int64 `json:"id"`
	Name string `json:"chatName"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
type messageResponse struct {
	ID int64 `json:"id"`
	Content string `json:"content"`
	IsAIResponse bool `json:"isAiResponse"`
	Timestamp string `json:"timestamp"`
	UserID *int64 `json:"userId,omitempty"`
}
type adminStatsResponse struct {
	Conversations int64 `json:"conversations"`
	Messages int64 `json:"messages"`
	UserMessages int64 `json:"user_messages"`
	AIMessages int64 `json:"ai_messages"`
	ActiveUsers int64 `json:"active_users"`
}

func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r); if !ok { http.Error(w, "unauthorized", 401); return }
	var req conversationRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "invalid request body", 400); return }
	}
	name := req.Name
	if strings.TrimSpace(name) == "" { name = req.ChatName }
	item, err := h.service.CreateConversation(r.Context(), userID, name)
	if err != nil { http.Error(w, "internal server error", 500); return }
	writeJSON(w, 201, toConversationResponse(item))
}
func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r); if !ok { http.Error(w, "unauthorized", 401); return }
	items, err := h.service.ListConversations(r.Context(), userID)
	if err != nil { http.Error(w, "internal server error", 500); return }
	response := make([]conversationResponse, 0, len(items))
	for _, item := range items { response = append(response, toConversationResponse(item)) }
	writeJSON(w, 200, response)
}
func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r); if !ok { http.Error(w, "unauthorized", 401); return }
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil { http.Error(w, "invalid chat id", 400); return }
	item, err := h.service.GetConversation(r.Context(), id, userID)
	if errors.Is(err, application.ErrConversationNotFound) { http.NotFound(w, r); return }
	if err != nil { http.Error(w, "internal server error", 500); return }
	writeJSON(w, 200, toConversationResponse(item))
}
func (h *Handler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r); if !ok { http.Error(w, "unauthorized", 401); return }
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil { http.Error(w, "invalid chat id", 400); return }
	if err := h.service.DeleteConversation(r.Context(), id, userID); errors.Is(err, application.ErrConversationNotFound) {
		http.NotFound(w, r); return
	} else if err != nil { http.Error(w, "internal server error", 500); return }
	w.WriteHeader(204)
}
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r); if !ok { http.Error(w, "unauthorized", 401); return }
	id, err := strconv.ParseInt(r.PathValue("chatId"), 10, 64)
	if err != nil { http.Error(w, "invalid chat id", 400); return }
	items, err := h.service.ListMessages(r.Context(), id, userID)
	if errors.Is(err, application.ErrConversationNotFound) { http.NotFound(w, r); return }
	if err != nil { http.Error(w, "internal server error", 500); return }
	response := make([]messageResponse, 0, len(items))
	for _, item := range items {
		response = append(response, messageResponse{ID:item.ID, Content:item.Content, IsAIResponse:item.IsAIResponse, Timestamp:item.CreatedAt.Format(time.RFC3339), UserID:item.UserID})
	}
	writeJSON(w, 200, response)
}
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r); if !ok { http.Error(w, "unauthorized", 401); return }
	chatID, err := strconv.ParseInt(r.URL.Query().Get("chatId"), 10, 64)
	if err != nil { http.Error(w, "invalid chatId", 400); return }
	content, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil || len(content) == 0 { http.Error(w, "invalid message", 400); return }
	message, err := h.service.SaveUserMessage(r.Context(), chatID, userID, string(content))
	if err != nil {
		if errors.Is(err, application.ErrConversationNotFound) { http.NotFound(w, r); return }
		http.Error(w, "internal server error", 500); return
	}
	writeJSON(w, 200, messageResponse{ID: message.ID, Content: message.Content, IsAIResponse: false, Timestamp: message.CreatedAt.Format(time.RFC3339), UserID: message.UserID})
}

func (h *Handler) DeepSeek(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r); if !ok { http.Error(w, "unauthorized", 401); return }
	chatID, err := strconv.ParseInt(r.URL.Query().Get("chatId"), 10, 64)
	if err != nil { http.Error(w, "invalid chatId", 400); return }
	input := strings.TrimSpace(r.URL.Query().Get("input"))
	if input == "" { http.Error(w, "input is empty", 400); return }

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher, ok := w.(http.Flusher)
	if !ok { http.Error(w, "streaming unsupported", 500); return }
	if err := h.service.StreamResponse(r.Context(), chatID, userID, input, func(chunk string) error {
		if _, err := w.Write([]byte(chunk)); err != nil { return err }
		flusher.Flush()
		return nil
	}); err != nil {
		if !errors.Is(err, application.ErrConversationNotFound) {
			return
		}
	}
}
func (h *Handler) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetAdminStats(r.Context())
	if err != nil { http.Error(w, "internal server error", 500); return }
	writeJSON(w, 200, adminStatsResponse{stats.Conversations, stats.Messages, stats.UserMessages, stats.AIMessages, stats.ActiveUsers})
}
func userID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.Header.Get("X-User-ID"), 10, 64)
	return id, err == nil && id > 0
}
func toConversationResponse(item domain.Conversation) conversationResponse {
	return conversationResponse{ID:item.ID, Name:item.Name, CreatedAt:item.CreatedAt.Format(time.RFC3339), UpdatedAt:item.UpdatedAt.Format(time.RFC3339)}
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(value)
}
