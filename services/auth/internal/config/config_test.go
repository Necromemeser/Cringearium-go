package config

import (
	"strings"
	"testing"
)

func TestAdministratorPasswordIsRequiredAndPreserved(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example/test")
	t.Setenv("JWT_SECRET", "configuration-test-secret")
	t.Setenv("JWT_TTL", "24h")
	for _, value := range []string{"", "       ", "short", strings.Repeat("x", 73)} {
		t.Setenv("ADMIN_PASSWORD", value)
		if _, err := Load(); err == nil {
			t.Fatalf("invalid ADMIN_PASSWORD of length %d accepted", len(value))
		}
	}
	value := " preserve leading and trailing spaces "
	t.Setenv("ADMIN_PASSWORD", value)
	cfg, err := Load()
	if err != nil || cfg.AdminPassword != value {
		t.Fatal("administrator password was changed or rejected", err)
	}
}
