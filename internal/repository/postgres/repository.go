package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ifaisalabid1/chat-app/internal/domain"
	"github.com/ifaisalabid1/chat-app/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db     *pgxpool.Pool
	logger *logger.Logger
}

func New(db *pgxpool.Pool, logger *logger.Logger) *Repository {
	return &Repository{db, logger}
}

func (r *Repository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
			INSERT INTO users (id, username, email, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now()
	user.ID = uuid.New()
	user.CreatedAt = now
	user.UpdatedAt = now

	args := []any{user.ID, user.Username, user.Email, user.CreatedAt, user.UpdatedAt}

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error(ctx, "failed to create user", err)
		return err
	}
	return nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
			SELECT id, username, email, created_at, updated_at
			FROM users
			WHERE id = $1
	`

	var user domain.User

	err := r.db.QueryRow(ctx, query, user.ID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.logger.Error(ctx, "failed to get user", err)
		return nil, err
	}

	return &user, nil
}

func (r *Repository) SaveMessage(ctx context.Context, msg *domain.Message) error {
	query := `
			INSERT INTO messages (id, room_id, user_id, username, content, type, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	msg.ID = uuid.New()
	msg.CreatedAt = now

	args := []any{msg.ID, msg.RoomID, msg.UserID, msg.Username, msg.Content, msg.Type, msg.CreatedAt}

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error(ctx, "failed to save message", err)
		return err
	}

	return nil
}

func (r *Repository) GetMessages(ctx context.Context, roomID string, limit, offset int) ([]*domain.Message, error) {
	query := `
			SELECT id, room_id, user_id, username, content, type, created_at
			FROM messages
			WHERE room_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, roomID, limit, offset)
	if err != nil {
		r.logger.Error(ctx, "failed to get messages", err)
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message

	for rows.Next() {
		var msg domain.Message

		err := rows.Scan(
			&msg.ID,
			&msg.RoomID,
			&msg.UserID,
			&msg.Username,
			&msg.Content,
			&msg.Type,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		messages = append(messages, &msg)
	}

	return messages, nil
}
