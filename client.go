package talky

import (
	"github.com/gorilla/websocket"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is the middleman between the websocket connection and the hub.
type Client struct {
	hub  *Hub
	user *User

	conn *websocket.Conn

	// Buffered channel for outbound messages.
	sendCh chan []byte
}

//func serveWs(hub *Hub, user *User, w http.ResponseWriter, r *http.Request)  {
//	conn, err := upgrader.Upgrade(w, r, nil)
//	if err
//}
