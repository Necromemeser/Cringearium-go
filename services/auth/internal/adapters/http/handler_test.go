package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Necromemeser/Cringearium-go/services/auth/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)

	users := mocks.NewMockUserRepository(ctrl)
	hasher := mocks.NewMockPasswordHasher(ctrl)
	tokens := mocks.NewMockTokenService(ctrl)

	auth := application.NewAuth(users, hasher, tokens)
	handler := NewHandler(auth)

	user := &domain.User{
		ID:           1,
		Username:     "test",
		Email:        "test@example.com",
		Role:         domain.RoleStudent,
		PasswordHash: "hashed",
	}

	users.
		EXPECT().
		FindByEmail(gomock.Any(), "test@example.com").
		Return(user, nil)

	hasher.
		EXPECT().
		Compare("hashed", "password123").
		Return(nil)

	tokens.
		EXPECT().
		Generate(gomock.Any(), user).
		Return("test-token", nil)

	body := bytes.NewBufferString(`{
		"email": "test@example.com",
		"password": "password123"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/login",
		body,
	)

	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Token != "test-token" {
		t.Fatalf("expected token %q, got %q", "test-token", response.Token)
	}
}
