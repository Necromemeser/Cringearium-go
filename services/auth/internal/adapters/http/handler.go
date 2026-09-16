package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Necromemeser/Cringearium-go/services/auth/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
)

type Handler struct {
	auth *application.Auth
}

func NewHandler(auth *application.Auth) *Handler {
	return &Handler{
		auth: auth,
	}
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type userResponse struct {
	ID             int64  `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	ProfileImageID string `json:"profile_image_id,omitempty"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.auth.Register(
		r.Context(),
		req.Username,
		req.Email,
		req.Password,
	)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidUsername),
			errors.Is(err, application.ErrInvalidEmail),
			errors.Is(err, application.ErrInvalidPassword):
			http.Error(w, err.Error(), http.StatusBadRequest)

		case errors.Is(err, application.ErrUserExists):
			http.Error(w, err.Error(), http.StatusConflict)

		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}

		return
	}

	response := registerResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	user, err := h.auth.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, application.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeUserResponse(w, user)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	username := r.URL.Query().Get("username")

	if email != "" && username != "" {
		http.Error(w, "only one query parameter is allowed", http.StatusBadRequest)
		return
	}

	if email == "" && username == "" {
		users, err := h.auth.GetAll(r.Context())
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		responses := make([]userResponse, 0, len(users))

		for _, user := range users {
			responses = append(responses, userResponse{
				ID:             user.ID,
				Username:       user.Username,
				Email:          user.Email,
				Role:           string(user.Role),
				ProfileImageID: valueOrEmpty(user.ProfileImageID),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(responses); err != nil {
			return
		}

		return
	}

	var (
		user *domain.User
		err  error
	)

	switch {
	case email != "":
		user, err = h.auth.GetByEmail(r.Context(), email)

	case username != "":
		user, err = h.auth.GetByUsername(r.Context(), username)
	}

	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidEmail),
			errors.Is(err, application.ErrInvalidUsername):
			http.Error(w, err.Error(), http.StatusBadRequest)

		case errors.Is(err, application.ErrUserNotFound):
			http.Error(w, "user not found", http.StatusNotFound)

		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}

		return
	}

	writeUserResponse(w, user)
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func writeUserResponse(w http.ResponseWriter, user *domain.User) {
	response := userResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}

	if user.ProfileImageID != nil {
		response.ProfileImageID = *user.ProfileImageID
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(response)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.auth.Login(
		r.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := loginResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	writeUserResponse(w, user)
}
