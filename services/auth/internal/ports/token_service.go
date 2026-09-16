package ports

import (
	"context"

	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
)

type TokenService interface {
	Generate(ctx context.Context, user *domain.User) (string, error)
	Validate(ctx context.Context, token string) (*domain.User, error)
}
