package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ifaisalabid1/chat-app/internal/domain"
	"github.com/ifaisalabid1/chat-app/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	client *redis.Client
	logger *logger.Logger
}

func New(client *redis.Client, logger *logger.Logger) *Repository {
	return &Repository{client, logger}
}

func (r *Repository) SaveMessage(ctx context.Context, roomID string, msg *domain.Message) error {
	key := fmt.Sprintf("room:%s:messages", roomID)

	data, err := json.Marshal(msg)
	if err != nil {
		r.logger.Error(ctx, "failed to marshal json", err)
		return err
	}

	err = r.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(msg.CreatedAt.UnixNano()),
		Member: data,
	}).Err()

	if err != nil {
		r.logger.Error(ctx, "failed to cache message", err)
		return err
	}

	r.client.Expire(ctx, key, 24*time.Hour)

	return nil
}

func (r *Repository) GetMessages(ctx context.Context, roomID string, limit int64) ([]*domain.Message, error) {
	key := fmt.Sprintf("room:%s:messages", roomID)

	results, err := r.client.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:   key,
		Start: 0,
		Stop:  limit - 1,
		Rev:   true,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		r.logger.Error(ctx, "failed to get cached messages", err)
		return nil, err
	}

	var messages []*domain.Message

	for _, result := range results {
		var msg domain.Message
		if err := json.Unmarshal([]byte(result), &msg); err != nil {
			continue
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

func (r *Repository) AddUserToRoom(ctx context.Context, roomID string, user *domain.User) error {
	key := fmt.Sprintf("room:%s:users", roomID)

	userData, err := json.Marshal(user)
	if err != nil {
		return err
	}

	if err := r.client.HSet(ctx, key, user.ID.String(), userData).Err(); err != nil {
		r.logger.Error(ctx, "failed to add user to room", err)
		return err
	}

	return nil
}

func (r *Repository) RemoveUserFromRoom(ctx context.Context, roomID string, user *domain.User) error {
	key := fmt.Sprintf("room:%s:users", roomID)

	if err := r.client.HDel(ctx, key, user.ID.String()).Err(); err != nil {
		r.logger.Error(ctx, "failed to delete user from room", err)
		return err
	}

	return nil
}

func (r *Repository) GetRoomUsers(ctx context.Context, roomID string) ([]*domain.User, error) {
	key := fmt.Sprintf("room:%s:users", roomID)

	results, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		r.logger.Error(ctx, "failed to get room users", err)
		return nil, err
	}

	var users []*domain.User

	for _, userData := range results {
		var user domain.User

		if err := json.Unmarshal([]byte(userData), &user); err != nil {
			continue
		}

		users = append(users, &user)
	}

	return users, nil
}
