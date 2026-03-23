package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/ifaisalabid1/chat-app/internal/domain"
	"github.com/ifaisalabid1/chat-app/internal/repository/postgres"
	"github.com/ifaisalabid1/chat-app/internal/repository/redis"
	"github.com/ifaisalabid1/chat-app/pkg/logger"
)

type ChatService struct {
	pgRepo    *postgres.Repository
	redisRepo *redis.Repository
	logger    *logger.Logger
}

func New(pgRepo *postgres.Repository, redisRepo *redis.Repository, logger *logger.Logger) *ChatService {
	return &ChatService{pgRepo, redisRepo, logger}
}

func (s *ChatService) CreateUser(ctx context.Context, username, email string) (*domain.User, error) {
	user := &domain.User{
		Username: username,
		Email:    email,
	}

	if err := s.pgRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *ChatService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.pgRepo.GetUserByID(ctx, id)
}

func (s *ChatService) SaveMessage(ctx context.Context, msg *domain.Message) error {
	if err := s.pgRepo.SaveMessage(ctx, msg); err != nil {
		return err
	}

	if err := s.redisRepo.SaveMessage(ctx, msg.RoomID, msg); err != nil {
		s.logger.Error(ctx, "failed to cache message", err)
	}

	return nil
}

func (s *ChatService) GetMessages(ctx context.Context, roomID string, limit, offset int) ([]*domain.Message, error) {
	if offset == 0 {
		cached, err := s.redisRepo.GetMessages(ctx, roomID, int64(limit))
		if err == nil && len(cached) > 0 {
			return cached, nil
		}
	}

	return s.pgRepo.GetMessages(ctx, roomID, limit, offset)
}

func (s *ChatService) AddUserToRoom(ctx context.Context, roomID string, user *domain.User) error {
	return s.redisRepo.AddUserToRoom(ctx, roomID, user)
}

func (s *ChatService) RemoveUserFromRoom(ctx context.Context, roomID string, user *domain.User) error {
	return s.redisRepo.RemoveUserFromRoom(ctx, roomID, user)
}

func (s *ChatService) GetRoomUsers(ctx context.Context, roomID string) ([]*domain.User, error) {
	return s.redisRepo.GetRoomUsers(ctx, roomID)
}
