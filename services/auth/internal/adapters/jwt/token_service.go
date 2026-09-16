package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
)

type TokenService struct {
	secret []byte
	ttl    time.Duration
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewTokenService(secret string, ttl time.Duration) *TokenService {
	return &TokenService{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (t *TokenService) Generate(
	_ context.Context,
	user *domain.User,
) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(t.secret)
}
