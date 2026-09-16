package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/mocks"
	"go.uber.org/mock/gomock"
)

var (
	errHash     = errors.New("hash error")
	errDatabase = errors.New("database error")
	errPassword = errors.New("wrong password")
	errToken    = errors.New("token error")
)

func TestAuth_Register(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*mocks.MockUserRepository, *mocks.MockPasswordHasher)
		username string
		email    string
		password string
		wantErr  error
	}{
		{
			name: "success",
			setup: func(users *mocks.MockUserRepository, hasher *mocks.MockPasswordHasher) {
				users.EXPECT().
					FindByEmail(gomock.Any(), "test@example.com").
					Return(nil, nil)

				users.EXPECT().
					FindByUsername(gomock.Any(), "test").
					Return(nil, nil)

				hasher.EXPECT().
					Hash("password123").
					Return("hashed-password", nil)

				users.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, user *domain.User) error {
						user.ID = 1
						return nil
					})
			},
			username: "test",
			email:    "test@example.com",
			password: "password123",
		},
		{
			name:     "invalid username",
			setup:    func(_ *mocks.MockUserRepository, _ *mocks.MockPasswordHasher) {},
			username: "",
			email:    "test@example.com",
			password: "password123",
			wantErr:  ErrInvalidUsername,
		},
		{
			name:     "invalid email",
			setup:    func(_ *mocks.MockUserRepository, _ *mocks.MockPasswordHasher) {},
			username: "test",
			email:    "",
			password: "password123",
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "invalid password",
			setup:    func(_ *mocks.MockUserRepository, _ *mocks.MockPasswordHasher) {},
			username: "test",
			email:    "test@example.com",
			password: "1234567",
			wantErr:  ErrInvalidPassword,
		},
		{
			name: "email already exists",
			setup: func(users *mocks.MockUserRepository, _ *mocks.MockPasswordHasher) {
				users.EXPECT().
					FindByEmail(gomock.Any(), "test@example.com").
					Return(&domain.User{ID: 1}, nil)
			},
			username: "test",
			email:    "test@example.com",
			password: "password123",
			wantErr:  ErrUserExists,
		},
		{
			name: "username already exists",
			setup: func(users *mocks.MockUserRepository, _ *mocks.MockPasswordHasher) {
				users.EXPECT().
					FindByEmail(gomock.Any(), "test@example.com").
					Return(nil, nil)

				users.EXPECT().
					FindByUsername(gomock.Any(), "test").
					Return(&domain.User{ID: 1}, nil)
			},
			username: "test",
			email:    "test@example.com",
			password: "password123",
			wantErr:  ErrUserExists,
		},
		{
			name: "hash error",
			setup: func(users *mocks.MockUserRepository, hasher *mocks.MockPasswordHasher) {
				users.EXPECT().
					FindByEmail(gomock.Any(), "test@example.com").
					Return(nil, nil)

				users.EXPECT().
					FindByUsername(gomock.Any(), "test").
					Return(nil, nil)

				hasher.EXPECT().
					Hash("password123").
					Return("", errHash)
			},
			username: "test",
			email:    "test@example.com",
			password: "password123",
			wantErr:  errHash,
		},
		{
			name: "repository error",
			setup: func(users *mocks.MockUserRepository, _ *mocks.MockPasswordHasher) {
				users.EXPECT().
					FindByEmail(gomock.Any(), "test@example.com").
					Return(nil, errDatabase)
			},
			username: "test",
			email:    "test@example.com",
			password: "password123",
			wantErr:  errDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			users := mocks.NewMockUserRepository(ctrl)
			hasher := mocks.NewMockPasswordHasher(ctrl)
			tokens := mocks.NewMockTokenService(ctrl)

			tt.setup(users, hasher)

			auth := NewAuth(users, hasher, tokens)

			user, err := auth.Register(
				context.Background(),
				tt.username,
				tt.email,
				tt.password,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if user.ID != 1 {
				t.Fatalf("expected user ID 1, got %d", user.ID)
			}

			if user.Username != "test" {
				t.Fatalf("expected username test, got %s", user.Username)
			}

			if user.Email != "test@example.com" {
				t.Fatalf("expected email test@example.com, got %s", user.Email)
			}

			if user.PasswordHash != "hashed-password" {
				t.Fatalf("expected hashed password, got %s", user.PasswordHash)
			}

			if user.Role != domain.RoleStudent {
				t.Fatalf("expected student role, got %s", user.Role)
			}
		})
	}
}

func TestAuth_Login(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*mocks.MockUserRepository, *mocks.MockPasswordHasher, *mocks.MockTokenService)
		email    string
		password string
		want     string
		wantErr  error
	}{
		{
			name: "success",
			setup: func(
				users *mocks.MockUserRepository,
				hasher *mocks.MockPasswordHasher,
				tokens *mocks.MockTokenService,
			) {
				user := &domain.User{
					ID:           1,
					Username:     "test",
					Email:        "test@example.com",
					PasswordHash: "hashed-password",
					Role:         domain.RoleStudent,
				}

				users.EXPECT().
					FindByEmail(gomock.Any(), "test@example.com").
					Return(user, nil)

				hasher.EXPECT().
					Compare("hashed-password", "password123").
					Return(nil)

				tokens.EXPECT().
					Generate(gomock.Any(), user).
					Return("jwt-token", nil)
			},
			email:    "test@example.com",
			password: "password123",
			want:     "jwt-token",
		},
		{
			name: "invalid credentials empty email",
			setup: func(
				_ *mocks.MockUserRepository,
				_ *mocks.MockPasswordHasher,
				_ *mocks.MockTokenService,
			) {
			},
			email:    "",
			password: "password123",
			wantErr:  ErrInvalidCredentials,
		},
		{
			name: "user not found",
			setup: func(
				users *mocks.MockUserRepository,
				_ *mocks.MockPasswordHasher,
				_ *mocks.MockTokenService,
			) {
				users.EXPECT().
					FindByEmail(gomock.Any(), "test@example.com").
					Return(nil, nil)
			},
			email:    "test@example.com",
			password: "password123",
			wantErr:  ErrInvalidCredentials,
		},
		{
			name: "wrong password",
			setup: func(
				users *mocks.MockUserRepository,
				hasher *mocks.MockPasswordHasher,
				_ *mocks.MockTokenService,
			) {
				user := &domain.User{
					ID:           1,
					Email:        "test@example.com",
					PasswordHash: "hashed-password",
				}

				users.EXPECT().
					FindByEmail(gomock.Any(), "test@example.com").
					Return(user, nil)

				hasher.EXPECT().
					Compare("hashed-password", "password123").
					Return(errPassword)
			},
			email:    "test@example.com",
			password: "password123",
			wantErr:  ErrInvalidCredentials,
		},
		{
			name: "token generation error",
			setup: func(
				users *mocks.MockUserRepository,
				hasher *mocks.MockPasswordHasher,
				tokens *mocks.MockTokenService,
			) {
				user := &domain.User{
					ID:           1,
					Email:        "test@example.com",
					PasswordHash: "hashed-password",
				}

				users.EXPECT().
					FindByEmail(gomock.Any(), "test@example.com").
					Return(user, nil)

				hasher.EXPECT().
					Compare("hashed-password", "password123").
					Return(nil)

				tokens.EXPECT().
					Generate(gomock.Any(), user).
					Return("", errToken)
			},
			email:    "test@example.com",
			password: "password123",
			wantErr:  errToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			users := mocks.NewMockUserRepository(ctrl)
			hasher := mocks.NewMockPasswordHasher(ctrl)
			tokens := mocks.NewMockTokenService(ctrl)

			tt.setup(users, hasher, tokens)

			auth := NewAuth(users, hasher, tokens)

			token, err := auth.Login(
				context.Background(),
				tt.email,
				tt.password,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if token != tt.want {
				t.Fatalf("expected token %q, got %q", tt.want, token)
			}
		})
	}
}

func TestAuth_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mocks.MockUserRepository)
		id      int64
		want    *domain.User
		wantErr error
	}{
		{
			name: "success",
			setup: func(users *mocks.MockUserRepository) {
				users.EXPECT().
					FindByID(gomock.Any(), int64(1)).
					Return(&domain.User{
						ID:       1,
						Username: "test",
						Email:    "test@example.com",
						Role:     domain.RoleStudent,
					}, nil)
			},
			id: 1,
			want: &domain.User{
				ID:       1,
				Username: "test",
				Email:    "test@example.com",
				Role:     domain.RoleStudent,
			},
		},
		{
			name: "user not found",
			setup: func(users *mocks.MockUserRepository) {
				users.EXPECT().
					FindByID(gomock.Any(), int64(1)).
					Return(nil, nil)
			},
			id:      1,
			wantErr: ErrUserNotFound,
		},
		{
			name: "repository error",
			setup: func(users *mocks.MockUserRepository) {
				users.EXPECT().
					FindByID(gomock.Any(), int64(1)).
					Return(nil, errDatabase)
			},
			id:      1,
			wantErr: errDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			users := mocks.NewMockUserRepository(ctrl)
			tt.setup(users)

			auth := NewAuth(
				users,
				mocks.NewMockPasswordHasher(ctrl),
				mocks.NewMockTokenService(ctrl),
			)

			user, err := auth.GetByID(
				context.Background(),
				tt.id,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if user.ID != tt.want.ID {
				t.Fatalf("expected ID %d, got %d", tt.want.ID, user.ID)
			}

			if user.Username != tt.want.Username {
				t.Fatalf("expected username %q, got %q", tt.want.Username, user.Username)
			}

			if user.Email != tt.want.Email {
				t.Fatalf("expected email %q, got %q", tt.want.Email, user.Email)
			}

			if user.Role != tt.want.Role {
				t.Fatalf("expected role %q, got %q", tt.want.Role, user.Role)
			}
		})
	}
}

func TestAuth_GetByEmail(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mocks.MockUserRepository)
		email   string
		want    *domain.User
		wantErr error
	}{
		{
			name: "success",
			setup: func(users *mocks.MockUserRepository) {
				users.EXPECT().
					FindByEmail(gomock.Any(), "test@example.com").
					Return(&domain.User{
						ID:       1,
						Username: "test",
						Email:    "test@example.com",
						Role:     domain.RoleStudent,
					}, nil)
			},
			email: "test@example.com",
			want: &domain.User{
				ID:       1,
				Username: "test",
				Email:    "test@example.com",
				Role:     domain.RoleStudent,
			},
		},
		{
			name:    "invalid email",
			setup:   func(_ *mocks.MockUserRepository) {},
			email:   "",
			wantErr: ErrInvalidEmail,
		},
		{
			name: "user not found",
			setup: func(users *mocks.MockUserRepository) {
				users.EXPECT().
					FindByEmail(gomock.Any(), "test@example.com").
					Return(nil, nil)
			},
			email:   "test@example.com",
			wantErr: ErrUserNotFound,
		},
		{
			name: "repository error",
			setup: func(users *mocks.MockUserRepository) {
				users.EXPECT().
					FindByEmail(gomock.Any(), "test@example.com").
					Return(nil, errDatabase)
			},
			email:   "test@example.com",
			wantErr: errDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			users := mocks.NewMockUserRepository(ctrl)
			tt.setup(users)

			auth := NewAuth(
				users,
				mocks.NewMockPasswordHasher(ctrl),
				mocks.NewMockTokenService(ctrl),
			)

			user, err := auth.GetByEmail(
				context.Background(),
				tt.email,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if user.ID != tt.want.ID {
				t.Fatalf("expected ID %d, got %d", tt.want.ID, user.ID)
			}

			if user.Username != tt.want.Username {
				t.Fatalf("expected username %q, got %q", tt.want.Username, user.Username)
			}

			if user.Email != tt.want.Email {
				t.Fatalf("expected email %q, got %q", tt.want.Email, user.Email)
			}

			if user.Role != tt.want.Role {
				t.Fatalf("expected role %q, got %q", tt.want.Role, user.Role)
			}
		})
	}
}

func TestAuth_GetByUsername(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*mocks.MockUserRepository)
		username string
		want     *domain.User
		wantErr  error
	}{
		{
			name: "success",
			setup: func(users *mocks.MockUserRepository) {
				users.EXPECT().
					FindByUsername(gomock.Any(), "test").
					Return(&domain.User{
						ID:       1,
						Username: "test",
						Email:    "test@example.com",
						Role:     domain.RoleStudent,
					}, nil)
			},
			username: "test",
			want: &domain.User{
				ID:       1,
				Username: "test",
				Email:    "test@example.com",
				Role:     domain.RoleStudent,
			},
		},
		{
			name:     "invalid username",
			setup:    func(_ *mocks.MockUserRepository) {},
			username: "",
			wantErr:  ErrInvalidUsername,
		},
		{
			name: "user not found",
			setup: func(users *mocks.MockUserRepository) {
				users.EXPECT().
					FindByUsername(gomock.Any(), "test").
					Return(nil, nil)
			},
			username: "test",
			wantErr:  ErrUserNotFound,
		},
		{
			name: "repository error",
			setup: func(users *mocks.MockUserRepository) {
				users.EXPECT().
					FindByUsername(gomock.Any(), "test").
					Return(nil, errDatabase)
			},
			username: "test",
			wantErr:  errDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			users := mocks.NewMockUserRepository(ctrl)
			tt.setup(users)

			auth := NewAuth(
				users,
				mocks.NewMockPasswordHasher(ctrl),
				mocks.NewMockTokenService(ctrl),
			)

			user, err := auth.GetByUsername(
				context.Background(),
				tt.username,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if user.ID != tt.want.ID {
				t.Fatalf("expected ID %d, got %d", tt.want.ID, user.ID)
			}

			if user.Username != tt.want.Username {
				t.Fatalf("expected username %q, got %q", tt.want.Username, user.Username)
			}

			if user.Email != tt.want.Email {
				t.Fatalf("expected email %q, got %q", tt.want.Email, user.Email)
			}

			if user.Role != tt.want.Role {
				t.Fatalf("expected role %q, got %q", tt.want.Role, user.Role)
			}
		})
	}
}

func TestAuth_GetAll(t *testing.T) {
	users := []*domain.User{
		{
			ID:       1,
			Username: "first",
			Email:    "first@example.com",
			Role:     domain.RoleStudent,
		},
		{
			ID:       2,
			Username: "second",
			Email:    "second@example.com",
			Role:     domain.RoleTeacher,
		},
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		repository := mocks.NewMockUserRepository(ctrl)

		repository.EXPECT().
			FindAll(gomock.Any()).
			Return(users, nil)

		auth := NewAuth(
			repository,
			mocks.NewMockPasswordHasher(ctrl),
			mocks.NewMockTokenService(ctrl),
		)

		got, err := auth.GetAll(context.Background())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != len(users) {
			t.Fatalf("expected %d users, got %d", len(users), len(got))
		}

		for i := range users {
			if got[i] != users[i] {
				t.Fatalf("expected user %v, got %v", users[i], got[i])
			}
		}
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		repository := mocks.NewMockUserRepository(ctrl)

		repository.EXPECT().
			FindAll(gomock.Any()).
			Return(nil, errDatabase)

		auth := NewAuth(
			repository,
			mocks.NewMockPasswordHasher(ctrl),
			mocks.NewMockTokenService(ctrl),
		)

		_, err := auth.GetAll(context.Background())

		if !errors.Is(err, errDatabase) {
			t.Fatalf("expected error %v, got %v", errDatabase, err)
		}
	})
}
