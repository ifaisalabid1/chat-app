package hub

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/ifaisalabid1/chat-app/internal/domain"
	"github.com/redis/go-redis/v9"
)

const redisPubSubChannel = "chat:messages"

type Hub struct {
	clients    map[string]map[*Client]bool
	broadcast  chan *domain.Message
	register   chan *Client
	unregister chan *Client
	rdb        *redis.Client
	logger     *slog.Logger
}

func NewHub(rdb *redis.Client, logger *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan *domain.Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rdb:        rdb,
		logger:     logger,
	}
}

func (h *Hub) Run(ctx context.Context) {
	pubsub := h.rdb.Subscribe(ctx, redisPubSubChannel)
	defer pubsub.Close()

	redisCh := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return

		case client := <-h.register:
			if _, ok := h.clients[client.RoomID]; !ok {
				h.clients[client.RoomID] = make(map[*Client]bool)
			}

			h.clients[client.RoomID][client] = true
			h.logger.Info("client registered", "user_id", client.UserID, "room_id", client.RoomID)

		case client := <-h.unregister:
			if room, ok := h.clients[client.RoomID]; ok {
				delete(room, client)
				close(client.Send)
			}
			h.logger.Info("client unregistered", "user_id", client.UserID, "room_id", client.RoomID)

		case msg := <-h.broadcast:
			data, err := json.Marshal(msg)
			if err != nil {
				h.logger.Error("failed to marshal", "error", err)
			}
			h.rdb.Publish(ctx, redisPubSubChannel, data)

		case redisMsg := <-redisCh:
			var msg domain.Message

			if err := json.Unmarshal([]byte(redisMsg.Payload), &msg); err != nil {
				h.logger.Error("failed to unmarshal", "error", err)
				continue
			}
			h.fanOut(&msg)
		}
	}
}

func (h *Hub) fanOut(msg *domain.Message) {
	roomID := msg.RoomID.String()

	clients, ok := h.clients[roomID]
	if !ok {
		return
	}

	event := domain.WSEvent{
		Type: "message",
		Payload: domain.MessagePayload{
			MessageID: msg.ID.String(),
			RoomID:    roomID,
			UserID:    msg.UserID.String(),
			Username:  msg.Username,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("failed to marshal", "error", err)
	}

	for client := range clients {
		select {
		case client.Send <- data:

		default:
			close(client.Send)
			delete(clients, client)
		}
	}
}

func (h *Hub) Register(c *Client) {
	h.register <- c
}

func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

func (h *Hub) Broadcast(msg *domain.Message) {
	h.broadcast <- msg
}
