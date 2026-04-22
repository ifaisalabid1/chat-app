package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	roomIDStr := r.URL.Query().Get("room_id")
	userIDStr := r.URL.Query().Get("user_id")

	// roomID, err := uuid.Parse(roomIDStr)
	// if err != nil {
	// 	http.Error(w, "invalid room_id", http.StatusBadRequest)
	// 	return
	// }

	// userID, err := uuid.Parse(userIDStr)
	// if err != nil {
	// 	http.Error(w, "invalid user_id", http.StatusBadRequest)
	// 	return
	// }

	user, room, err := h.hub.usecase.JoinOrCreateRoom(r.Context(), userIDStr, roomIDStr)
	if err != nil {
		slog.Error("failed to join or create room", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("failed to upgrade to websocket", "error", err)
		return
	}

	client := &Client{
		hub:    h.hub,
		conn:   conn,
		userID: user.ID,
		roomID: room.ID,
		send:   make(chan []byte, 256),
	}

	h.hub.register <- client

	history, err := h.hub.usecase.GetRoomHistory(r.Context(), room.ID)
	if err == nil {
		for _, msg := range history {
			payload, err := json.Marshal(msg)
			if err != nil {
				slog.Error("failed to marshal json", "error", err)
				return
			}
			client.send <- payload
		}
	}

	go client.writePump()
	go client.readPump()
}
