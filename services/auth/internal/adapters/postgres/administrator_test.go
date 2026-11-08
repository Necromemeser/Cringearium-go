package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Necromemeser/Cringearium-go/internal/testsupport"
	password "github.com/Necromemeser/Cringearium-go/services/auth/internal/adapters/bcrypt"
	token "github.com/Necromemeser/Cringearium-go/services/auth/internal/adapters/jwt"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
)

func TestAdministratorBootstrapCreatesAndRotatesExistingCredentials(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprintf("legacy=%t", legacy), func(t *testing.T) {
			conn := testsupport.Postgres(t)
			for _, file := range []string{"000001_initial_auth.up.sql", "000002_seed_admin.up.sql"} {
				raw, err := migrationFiles.ReadFile("migrations/" + file)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := conn.Exec(string(raw)); err != nil {
					t.Fatal(err)
				}
			}
			var count int
			if err := conn.Get(&count, `SELECT count(*) FROM users`); err != nil || count != 0 {
				t.Fatal("static seed must not create a default administrator", count, err)
			}
			hasher := password.NewHasher(12)
			previousHash, err := hasher.Hash("previous-administrator-password")
			if err != nil {
				t.Fatal(err)
			}
			if legacy {
				if _, err := conn.Exec(`INSERT INTO users(username,email,password_hash,role) VALUES('legacy-admin','admin@mail.ru',$1,'admin')`, previousHash); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := conn.Exec(`INSERT INTO users(username,email,password_hash,role) VALUES('unrelated','unrelated@example.test',$1,'student')`, previousHash); err != nil {
				t.Fatal(err)
			}
			repo := NewUserRepository(&DB{conn: conn})
			ctx := context.Background()
			tokens := token.NewTokenService("bootstrap-test-signing-secret", time.Hour)
			service := application.NewAuth(repo, hasher, tokens)
			if err := application.BootstrapAdministrator(ctx, repo, hasher, ""); err == nil {
				t.Fatal("missing administrator password accepted")
			}
			configured := "configured-administrator-password"
			if err := application.BootstrapAdministrator(ctx, repo, hasher, configured); err != nil {
				t.Fatal(err)
			}
			admin, err := repo.FindByEmail(ctx, "admin@mail.ru")
			if err != nil || admin == nil || admin.Role != domain.RoleAdmin {
				t.Fatal("administrator missing", err)
			}
			if legacy && admin.Username != "legacy-admin" {
				t.Fatal("existing username changed")
			}
			jwt, err := service.Login(ctx, "admin@mail.ru", configured)
			if err != nil {
				t.Fatal("configured password cannot log in", err)
			}
			identity, err := tokens.Validate(ctx, jwt)
			if err != nil || identity.ID != admin.ID || identity.Role != domain.RoleAdmin {
				t.Fatal("invalid administrator token", err)
			}
			if _, err := service.Login(ctx, "admin@mail.ru", "previous-administrator-password"); !errors.Is(err, application.ErrInvalidCredentials) {
				t.Fatal("previous credentials still accepted", err)
			}
			if err := application.BootstrapAdministrator(ctx, repo, hasher, configured); err != nil {
				t.Fatal(err)
			}
			same, err := repo.FindByEmail(ctx, "admin@mail.ru")
			if err != nil || same.ID != admin.ID || same.PasswordHash != admin.PasswordHash {
				t.Fatal("restart reset existing identity/hash", err)
			}
			rotated := "rotated-administrator-password"
			if err := application.BootstrapAdministrator(ctx, repo, hasher, rotated); err != nil {
				t.Fatal(err)
			}
			if _, err := service.Login(ctx, "admin@mail.ru", configured); !errors.Is(err, application.ErrInvalidCredentials) {
				t.Fatal("rotation did not replace password", err)
			}
			if _, err := service.Login(ctx, "admin@mail.ru", rotated); err != nil {
				t.Fatal("rotated password cannot log in", err)
			}
			other, err := repo.FindByEmail(ctx, "unrelated@example.test")
			if err != nil || other.PasswordHash != previousHash || other.Role != domain.RoleStudent {
				t.Fatal("unrelated account altered", err)
			}
			raw, err := migrationFiles.ReadFile("migrations/000002_seed_admin.down.sql")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := conn.Exec(string(raw)); err != nil {
				t.Fatal(err)
			}
			admin, err = repo.FindByEmail(ctx, "admin@mail.ru")
			if err != nil || admin == nil || admin.ID != identity.ID {
				t.Fatal("rollback removed configured administrator", err)
			}
		})
	}
}
