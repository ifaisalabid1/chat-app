package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/ifaisalabid1/chat-app/internal/domain"
	"github.com/redis/go-redis/v9"
)

type roomMessage struct {
	roomID  uuid.UUID
	payload []byte
}

type Hub struct {
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan roomMessage
	redis      *redis.Client
	usecase    domain.ChatUsecase
}

func NewHub(rdb *redis.Client, uc domain.ChatUsecase) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan roomMessage, 256),
		redis:      rdb,
		usecase:    uc,
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case client := <-h.register:
			h.clients[client] = true
			if _, ok := h.rooms[client.roomID.String()]; !ok {
				h.rooms[client.roomID.String()] = make(map[*Client]bool)

				go h.subscribeToRoom(ctx, client.roomID)
			}
			h.rooms[client.roomID.String()][client] = true
			slog.Info("client registered", "user_id", client.userID, "room_id", client.roomID)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.rooms[client.roomID.String()], client)
				close(client.send)

				slog.Info("client unregistered", "user_id", client.userID)

				if len(h.rooms[client.roomID.String()]) == 0 {
					delete(h.rooms, client.roomID.String())
				}
			}

		case msg := <-h.broadcast:
			if roomClients, ok := h.rooms[msg.roomID.String()]; ok {
				for client := range roomClients {
					select {
					case client.send <- msg.payload:
					default:
						close(client.send)
						delete(h.clients, client)
						delete(h.rooms[msg.roomID.String()], client)
					}
				}
			}
		}
	}
}

func (h *Hub) PublishMessage(ctx context.Context, roomID, userID uuid.UUID, content string) error {
	msg, err := h.usecase.SendMessage(ctx, roomID, userID, content)
	if err != nil {
		return fmt.Errorf("failed to publish message to redis: %w", err)
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	return h.redis.Publish(ctx, "room:"+roomID.String(), payload).Err()
}

func (h *Hub) subscribeToRoom(ctx context.Context, roomID uuid.UUID) {
	pubsub := h.redis.Subscribe(ctx, "room:"+roomID.String())
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-ch:
			h.broadcast <- roomMessage{
				roomID:  roomID,
				payload: []byte(msg.Payload),
			}
		}
	}
}
