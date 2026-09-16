package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestMiddleware_Auth(t *testing.T) {
	user := &domain.User{
		ID:       42,
		Username: "test",
		Role:     domain.RoleStudent,
	}

	tests := []struct {
		name           string
		authorization  string
		validateResult *domain.User
		validateError  error
		wantStatus     int
		wantUser       bool
		validateCalled bool
	}{
		{
			name:           "valid token",
			authorization:  "Bearer valid-token",
			validateResult: user,
			wantStatus:     http.StatusOK,
			wantUser:       true,
			validateCalled: true,
		},
		{
			name:           "lowercase bearer",
			authorization:  "bearer valid-token",
			validateResult: user,
			wantStatus:     http.StatusOK,
			wantUser:       true,
			validateCalled: true,
		},
		{
			name:           "missing authorization",
			authorization:  "",
			wantStatus:     http.StatusUnauthorized,
			wantUser:       false,
			validateCalled: false,
		},
		{
			name:           "invalid authorization format",
			authorization:  "valid-token",
			wantStatus:     http.StatusUnauthorized,
			wantUser:       false,
			validateCalled: false,
		},
		{
			name:           "too many authorization parts",
			authorization:  "Bearer valid-token extra",
			wantStatus:     http.StatusUnauthorized,
			wantUser:       false,
			validateCalled: false,
		},
		{
			name:           "invalid token",
			authorization:  "Bearer invalid-token",
			validateError:  context.Canceled,
			wantStatus:     http.StatusUnauthorized,
			wantUser:       false,
			validateCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tokenService := mocks.NewMockTokenService(ctrl)

			if tt.validateCalled {
				tokenService.
					EXPECT().
					Validate(gomock.Any(), gomock.Any()).
					Return(tt.validateResult, tt.validateError)
			}

			middleware := NewMiddleware(tokenService)

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotUser, ok := UserFromContext(r.Context())

				if ok != tt.wantUser {
					t.Fatalf(
						"expected user presence %v, got %v",
						tt.wantUser,
						ok,
					)
				}

				if tt.wantUser && gotUser != user {
					t.Fatalf("unexpected user in context")
				}

				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.Auth(next)

			req := httptest.NewRequest(
				http.MethodGet,
				"/test",
				nil,
			)

			if tt.authorization != "" {
				req.Header.Set(
					"Authorization",
					tt.authorization,
				)
			}

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.wantStatus,
					rec.Code,
				)
			}
		})
	}
}
