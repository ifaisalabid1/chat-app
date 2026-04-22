package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ifaisalabid1/chat-app/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresChatRepo struct {
	db *pgxpool.Pool
}

func NewPostgresChatRepo(db *pgxpool.Pool) domain.ChatRepository {
	return &postgresChatRepo{db: db}
}

func (r *postgresChatRepo) GetOrCreateUser(ctx context.Context, username string) (*domain.User, error) {
	query := `
			WITH new_user AS (
				INSERT INTO users (username)
				VALUES ($1)
				ON CONFLICT (username) DO NOTHING
				RETURNING id, username, created_at
			)
			SELECT id, username, created_at FROM new_user
			UNION ALL
			SELECT id, username, created_at FROM users
			WHERE username = $1
			LIMIT 1;
	`

	var user domain.User

	if err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *postgresChatRepo) GetOrCreateRoom(ctx context.Context, name string) (*domain.Room, error) {
	query := `
			WITH new_room AS (
				INSERT INTO rooms (name) VALUES ($1)
				ON CONFLICT (name) DO NOTHING
				RETURNING id, name, created_at
			)
			SELECT id, name, created_at FROM new_room
			UNION ALL
			SELECT id, name, created_at FROM rooms
			WHERE name = $1
			LIMIT 1;
	`

	var room domain.Room

	if err := r.db.QueryRow(ctx, query, name).Scan(
		&room.ID,
		&room.Name,
		&room.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &room, nil
}

func (r *postgresChatRepo) SaveMessage(ctx context.Context, msg *domain.Message) error {
	query := `
			INSERT INTO messages (room_id, user_id, content)
			VALUES ($1, $2, $3)
			RETURNING id, created_at
	`

	return r.db.QueryRow(ctx, query, msg.RoomID, msg.UserID, msg.Content).Scan(
		&msg.ID,
		&msg.CreatedAt,
	)
}

func (r *postgresChatRepo) GetRecentMessages(ctx context.Context, roomID uuid.UUID, limit int) ([]domain.Message, error) {
	query := `
			SELECT id, room_id, user_id, content, created_at
			FROM messages
			WHERE room_id = $1
			ORDER BY created_at DESC
			LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, roomID, limit)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message

	for rows.Next() {
		var msg domain.Message

		if err := rows.Scan(
			&msg.ID,
			&msg.RoomID,
			&msg.UserID,
			&msg.Content,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return messages, nil
}
