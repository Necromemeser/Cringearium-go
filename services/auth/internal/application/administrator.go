package application

import (
	"context"
	"errors"
	"strings"

	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/ports"
)

const AdministratorEmail = "admin@mail.ru"

// BootstrapAdministrator makes the supplied password authoritative on both new and existing installations.
// No default password is provided; unchanged credentials retain their existing salted hash.
func BootstrapAdministrator(ctx context.Context, users ports.AdministratorRepository, hasher ports.PasswordHasher, password string) error {
	if strings.TrimSpace(password) == "" || len(password) < 8 || len(password) > 72 {
		return errors.New("ADMIN_PASSWORD must contain between 8 and 72 bytes")
	}
	user, err := users.FindByEmail(ctx, AdministratorEmail)
	if err != nil {
		return err
	}
	if user != nil && user.Role == domain.RoleAdmin && hasher.Compare(user.PasswordHash, password) == nil {
		return nil
	}
	hash, err := hasher.Hash(password)
	if err != nil {
		return err
	}
	return users.EnsureAdministrator(ctx, hash)
}
