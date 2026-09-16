package ports

import (
	"context"

	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindAll(ctx context.Context) ([]*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
}
