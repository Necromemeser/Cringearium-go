package ports

import (
	"context"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
)

// AdministratorRepository supports configuring the reserved administrator at startup.
type AdministratorRepository interface {
	FindByEmail(context.Context, string) (*domain.User, error)
	EnsureAdministrator(context.Context, string) error
}
