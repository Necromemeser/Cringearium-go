package ports

import (
	"context"

	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
)

// AuthUseCases is the input port used by transport adapters.
type AuthUseCases interface {
	Register(
		ctx context.Context,
		username string,
		email string,
		password string,
	) (*domain.User, error)
	GetAll(ctx context.Context) ([]*domain.User, error)
	GetByID(
		ctx context.Context,
		id int64,
	) (*domain.User, error)
	GetByEmail(
		ctx context.Context,
		email string,
	) (*domain.User, error)
	GetByUsername(
		ctx context.Context,
		username string,
	) (*domain.User, error)
	Login(
		ctx context.Context,
		email string,
		password string,
	) (string, error)
}
