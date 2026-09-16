package application

import (
	"context"
	"errors"
	"strings"

	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/ports"
)

var (
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Auth struct {
	users  ports.UserRepository
	hasher ports.PasswordHasher
	tokens ports.TokenService
}

func NewAuth(
	users ports.UserRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenService,
) *Auth {
	return &Auth{
		users:  users,
		hasher: hasher,
		tokens: tokens,
	}
}

func (a *Auth) Register(
	ctx context.Context,
	username string,
	email string,
	password string,
) (*domain.User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(strings.ToLower(email))

	if username == "" {
		return nil, ErrInvalidUsername
	}

	if email == "" {
		return nil, ErrInvalidEmail
	}

	if len(password) < 8 {
		return nil, ErrInvalidPassword
	}

	existing, err := a.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, ErrUserExists
	}

	existing, err = a.users.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, ErrUserExists
	}

	passwordHash, err := a.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         domain.RoleStudent,
	}

	if err := a.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (a *Auth) GetAll(ctx context.Context) ([]*domain.User, error) {
	return a.users.FindAll(ctx)
}

func (a *Auth) GetByID(
	ctx context.Context,
	id int64,
) (*domain.User, error) {
	user, err := a.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (a *Auth) GetByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" {
		return nil, ErrInvalidEmail
	}

	user, err := a.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (a *Auth) GetByUsername(
	ctx context.Context,
	username string,
) (*domain.User, error) {
	username = strings.TrimSpace(username)

	if username == "" {
		return nil, ErrInvalidUsername
	}

	user, err := a.users.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	user, err := a.users.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", ErrInvalidCredentials
	}

	if err := a.hasher.Compare(user.PasswordHash, password); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := a.tokens.Generate(ctx, user)
	if err != nil {
		return "", err
	}

	return token, nil
}
