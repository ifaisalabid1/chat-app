package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ifaisalabid1/chat-app/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepository struct {
	pool *pgxpool.Pool
}

func NewMessageRepository(pool *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{pool: pool}
}

func (r *MessageRepository) Save(ctx context.Context, msg *domain.Message) error {
	query := `
			INSERT INTO messages (id, room_id, user_id, content, created_at)
			VALUES ($1, $2, $3, $4, $5)
	`

	args := []any{msg.ID, msg.RoomID, msg.UserID, msg.Content, msg.CreatedAt}

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}

func (r *MessageRepository) GetByRoom(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]domain.Message, error) {
	query := `
			SELECT m.id, m.room_id, m.user_id, u.username, m.content, m.created_at
			FROM messages m
			JOIN users u ON u.id = m.user_id
			WHERE m.room_id = $1
			ORDER BY m.created_at DESC
			LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var msgs []domain.Message

	for rows.Next() {
		var msg domain.Message

		if err := rows.Scan(&msg.ID,
			&msg.RoomID,
			&msg.UserID,
			&msg.Username,
			&msg.Content,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msgs = append(msgs, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return msgs, nil
}
