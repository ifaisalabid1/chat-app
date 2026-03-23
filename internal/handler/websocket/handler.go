package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/ifaisalabid1/chat-app/internal/domain"
)

type Client struct {
	ID     string
	User   *domain.User
	RoomID string
	Conn   *websocket.Conn
	Send   chan *domain.WSMessage
}
