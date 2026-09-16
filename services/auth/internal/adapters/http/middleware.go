package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/ports"
)

type contextKey string

const userContextKey contextKey = "auth_user"

type Middleware struct {
	tokens ports.TokenService
}

func NewMiddleware(tokens ports.TokenService) *Middleware {
	return &Middleware{
		tokens: tokens,
	}
}

func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := m.tokens.Validate(r.Context(), parts[1])
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserFromContext(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(userContextKey).(*domain.User)
	return user, ok
}
