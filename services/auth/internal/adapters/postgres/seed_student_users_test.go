package postgres

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Necromemeser/Cringearium-go/internal/testsupport"
	password "github.com/Necromemeser/Cringearium-go/services/auth/internal/adapters/bcrypt"
	token "github.com/Necromemeser/Cringearium-go/services/auth/internal/adapters/jwt"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
)

func TestStudentSeedsLoginAndConflictSafeRollback(t *testing.T) {
	conn := testsupport.Postgres(t)
	db := &DB{conn: conn, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	repo := NewUserRepository(db)
	ctx := context.Background()
	users, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 51 {
		t.Fatalf("want admin and 50 students, got %d", len(users))
	}
	salts := map[string]bool{}
	var first *domain.User
	for i := 1; i <= 50; i++ {
		user, err := repo.FindByEmail(ctx, fmt.Sprintf("user%d@mail.ru", i))
		if err != nil || user == nil {
			t.Fatal("missing seed", i, err)
		}
		if user.Username != fmt.Sprintf("user%d", i) || user.Role != domain.RoleStudent || user.PasswordHash == "12345678" {
			t.Fatal("invalid seeded account", i)
		}
		salts[user.PasswordHash] = true
		if i == 1 {
			first = user
		}
	}
	if len(salts) != 50 {
		t.Fatal("password hashes must use individual salts")
	}
	tokens := token.NewTokenService("seed-integration-test-secret", time.Hour)
	service := application.NewAuth(repo, password.NewHasher(12), tokens)
	for _, i := range []int{1, 50} {
		jwt, err := service.Login(ctx, fmt.Sprintf("user%d@mail.ru", i), "12345678")
		if err != nil {
			t.Fatal("seed login failed", i, err)
		}
		user, err := tokens.Validate(ctx, jwt)
		if err != nil || user.Username != fmt.Sprintf("user%d", i) || user.Role != domain.RoleStudent {
			t.Fatal("invalid seed token", i, err)
		}
	}
	if _, err := service.Login(ctx, "user1@mail.ru", "wrong-password"); !errors.Is(err, application.ErrInvalidCredentials) {
		t.Fatal("wrong password accepted", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatal("restarting migrations failed", err)
	}
	again, err := repo.FindByEmail(ctx, "user1@mail.ru")
	if err != nil || again.ID != first.ID || again.PasswordHash != first.PasswordHash {
		t.Fatal("restart recreated accounts or reset credentials", err)
	}
	apply := func(direction string) {
		t.Helper()
		sql, err := migrationFiles.ReadFile("migrations/000003_seed_student_users." + direction + ".sql")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := conn.Exec(string(sql)); err != nil {
			t.Fatal(direction, err)
		}
	}
	apply("down")
	// Conflicts on either unique field must preserve the existing user and their role/password.
	if _, err := conn.Exec(`INSERT INTO users(username,email,password_hash,role) VALUES('user1','original@example.test',$1,'teacher'),('original-user','user2@mail.ru',$1,'student')`, first.PasswordHash); err != nil {
		t.Fatal(err)
	}
	apply("up")
	var count int
	if err := conn.Get(&count, `SELECT count(*) FROM seeded_student_users`); err != nil || count != 48 {
		t.Fatal("conflicts not skipped", count, err)
	}
	preserved, err := repo.FindByUsername(ctx, "user1")
	if err != nil || preserved.Email != "original@example.test" || preserved.Role != domain.RoleTeacher || preserved.PasswordHash != first.PasswordHash {
		t.Fatal("existing account overwritten", err)
	}
	apply("down")
	users, err = repo.FindAll(ctx)
	if err != nil || len(users) != 3 {
		t.Fatal("rollback removed unrelated accounts", len(users), err)
	}
}
