package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ifaisalabid1/chat-app/internal/domain"
)

type chatUsecase struct {
	repo           domain.ChatRepository
	contextTimeout time.Duration
}

func NewChatUsecase(repo domain.ChatRepository, timeout time.Duration) domain.ChatUsecase {
	return &chatUsecase{
		repo:           repo,
		contextTimeout: timeout,
	}
}

func (u *chatUsecase) JoinOrCreateRoom(ctx context.Context, username, roomName string) (*domain.User, *domain.Room, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	user, err := u.repo.GetOrCreateUser(ctx, username)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get/create user: %w", err)
	}

	room, err := u.repo.GetOrCreateRoom(ctx, roomName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get/create room: %w", err)
	}

	return user, room, nil
}

func (u *chatUsecase) SendMessage(ctx context.Context, roomID, userID uuid.UUID, content string) (*domain.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if content == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	msg := &domain.Message{
		RoomID:  roomID.String(),
		UserID:  userID.String(),
		Content: content,
	}

	if err := u.repo.SaveMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	return msg, nil
}

func (u *chatUsecase) GetRoomHistory(ctx context.Context, roomID uuid.UUID) ([]domain.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	return u.repo.GetRecentMessages(ctx, roomID, 25)
}
