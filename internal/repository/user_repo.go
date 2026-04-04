package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ifaisalabid1/chat-app/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
			INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
	`

	args := []any{user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt}

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
			SELECT id, username, email, password_hash, created_at, updated_at
			FROM users
			WHERE email = $1
	`

	var user *domain.User

	if err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}
