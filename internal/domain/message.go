package domain

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID
	RoomID    uuid.UUID
	UserID    uuid.UUID
	Username  string
	Content   string
	CreatedAt time.Time
}

type WSEvent struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type MessagePayload struct {
	MessageID string `json:"message_id"`
	RoomID    string `json:"room_id"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}
