package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key"

func TestTokenService_GenerateAndValidate(t *testing.T) {
	service := NewTokenService(
		testSecret,
		time.Hour,
	)

	user := &domain.User{
		ID:       42,
		Username: "test",
		Role:     domain.RoleStudent,
	}

	token, err := service.Generate(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Fatal("expected token, got empty string")
	}

	got, err := service.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if got.ID != user.ID {
		t.Fatalf("expected ID %d, got %d", user.ID, got.ID)
	}

	if got.Username != user.Username {
		t.Fatalf("expected username %q, got %q", user.Username, got.Username)
	}

	if got.Role != user.Role {
		t.Fatalf("expected role %q, got %q", user.Role, got.Role)
	}
}

func TestTokenService_ValidateExpiredToken(t *testing.T) {
	service := NewTokenService(
		testSecret,
		time.Hour,
	)

	now := time.Now()

	claims := Claims{
		UserID:   42,
		Username: "test",
		Role:     string(domain.RoleStudent),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "test",
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	tokenString, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = service.Validate(
		context.Background(),
		tokenString,
	)

	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestTokenService_ValidateInvalidSignature(t *testing.T) {
	service := NewTokenService(
		testSecret,
		time.Hour,
	)

	user := &domain.User{
		ID:       42,
		Username: "test",
		Role:     domain.RoleStudent,
	}

	token, err := service.Generate(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	serviceWithWrongSecret := NewTokenService(
		"wrong-secret",
		time.Hour,
	)

	_, err = serviceWithWrongSecret.Validate(
		context.Background(),
		token,
	)

	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestTokenService_ValidateWrongAlgorithm(t *testing.T) {
	service := NewTokenService(
		testSecret,
		time.Hour,
	)

	claims := Claims{
		UserID:   42,
		Username: "test",
		Role:     string(domain.RoleStudent),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "test",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS512,
		claims,
	)

	tokenString, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = service.Validate(
		context.Background(),
		tokenString,
	)

	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestTokenService_ValidateMalformedToken(t *testing.T) {
	service := NewTokenService(
		testSecret,
		time.Hour,
	)

	_, err := service.Validate(
		context.Background(),
		"not-a-jwt-token",
	)

	if err == nil {
		t.Fatal("expected validation error")
	}
}
