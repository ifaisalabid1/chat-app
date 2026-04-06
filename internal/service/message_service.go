package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ifaisalabid1/chat-app/internal/domain"
	"github.com/ifaisalabid1/chat-app/internal/hub"
	"github.com/ifaisalabid1/chat-app/internal/repository"
)

type MessageService struct {
	messages *repository.MessageRepository
	hub      *hub.Hub
}

func NewMessageService(messages *repository.MessageRepository, hub *hub.Hub) *MessageService {
	return &MessageService{
		messages: messages,
		hub:      hub,
	}
}

func (s *MessageService) Send(ctx context.Context, roomID, userID uuid.UUID, username, content string) (*domain.Message, error) {
	if len(content) == 0 || len(content) > 4000 {
		return nil, fmt.Errorf("message length must be 1-4000 chars")
	}

	msg := &domain.Message{
		ID:        uuid.New(),
		RoomID:    roomID,
		UserID:    userID,
		Username:  username,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.messages.Save(ctx, msg); err != nil {
		return nil, err
	}

	s.hub.Broadcast(msg)
	return msg, nil
}
