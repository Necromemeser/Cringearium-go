package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Necromemeser/Cringearium-go/services/auth/internal/domain"
)

type UserRepository struct {
	db *DB
}

type userRow struct {
	ID             int64   `db:"id"`
	Username       string  `db:"username"`
	Email          string  `db:"email"`
	PasswordHash   string  `db:"password_hash"`
	Role           string  `db:"role"`
	ProfileImageID *string `db:"profile_image_id"`
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(
	ctx context.Context,
	user *domain.User,
) error {
	query := `
		INSERT INTO users (
			username,
			email,
			password_hash,
			role,
			profile_image_id
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	return r.db.conn.GetContext(
		ctx,
		&user.ID,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.ProfileImageID,
	)
}

func (r *UserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	query := `
		SELECT
			id,
			username,
			email,
			password_hash,
			role,
			profile_image_id
		FROM users
		WHERE email = $1
	`

	return r.findOne(ctx, query, email)
}

func (r *UserRepository) FindByID(
	ctx context.Context,
	id int64,
) (*domain.User, error) {
	query := `
			SELECT
				id,
				username,
				email,
				password_hash,
				role,
				profile_image_id
			FROM users
			WHERE id = $1
		`

	return r.findOne(ctx, query, id)
}

func (r *UserRepository) FindByUsername(
	ctx context.Context,
	username string,
) (*domain.User, error) {
	query := `
		SELECT
			id,
			username,
			email,
			password_hash,
			role,
			profile_image_id
		FROM users
		WHERE username = $1
	`

	return r.findOne(ctx, query, username)
}

func (r *UserRepository) findOne(
	ctx context.Context,
	query string,
	arg any,
) (*domain.User, error) {
	var row userRow

	err := r.db.conn.GetContext(ctx, &row, query, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &domain.User{
		ID:             row.ID,
		Username:       row.Username,
		Email:          row.Email,
		PasswordHash:   row.PasswordHash,
		Role:           domain.Role(row.Role),
		ProfileImageID: row.ProfileImageID,
	}, nil
}

func (r *UserRepository) FindAll(ctx context.Context) ([]*domain.User, error) {
	query := `
		SELECT
			id,
			username,
			email,
			password_hash,
			role,
			profile_image_id
		FROM users
		ORDER BY id
	`

	var rows []userRow

	if err := r.db.conn.SelectContext(ctx, &rows, query); err != nil {
		return nil, err
	}

	users := make([]*domain.User, 0, len(rows))

	for _, row := range rows {
		users = append(users, &domain.User{
			ID:             row.ID,
			Username:       row.Username,
			Email:          row.Email,
			PasswordHash:   row.PasswordHash,
			Role:           domain.Role(row.Role),
			ProfileImageID: row.ProfileImageID,
		})
	}

	return users, nil
}
