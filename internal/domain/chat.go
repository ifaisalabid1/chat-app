package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type Room struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID        uuid.UUID `json:"id"`
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatRepository interface {
	GetOrCreateUser(ctx context.Context, username string) (*User, error)
	GetOrCreateRoom(ctx context.Context, name string) (*Room, error)
	SaveMessage(ctx context.Context, msg *Message) error
	GetRecentMessages(ctx context.Context, roomID uuid.UUID, limit int) ([]Message, error)
}

type ChatUsecase interface {
	JoinOrCreateRoom(ctx context.Context, username, roomName string) (*User, *Room, error)
	SendMessage(ctx context.Context, roomID, userID uuid.UUID, content string) (*Message, error)
	GetRoomHistory(ctx context.Context, roomID uuid.UUID) ([]Message, error)
}
