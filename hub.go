package talky

import "log"

// BroadcastMessage defines the structure of the broadcasts sent by client websocket connection.
type BroadcastMessage struct {
	RoomID string `json:"room_id"`
	UserID uint   `json:"user_id"`

	Payload interface{} `json:"payload"`
}

type Hub struct {
	rooms   map[string]*Room
	clients map[uint]*Client

	registerCh   chan *Client
	unregisterCh chan *Client
	broadcastCh  chan *BroadcastMessage
}

func NewHub() *Hub {
	return &Hub{
		rooms:        make(map[string]*Room),
		clients:      make(map[uint]*Client),
		registerCh:   make(chan *Client),
		unregisterCh: make(chan *Client),
		broadcastCh:  make(chan *BroadcastMessage),
	}
}

func (h *Hub) run() {
	conferencing := 1
	log.Printf("%v", conferencing)
	for {
		select {
		case client := <-h.registerCh:
			h.clients[client.user.ID] = client
		case client := <-h.unregisterCh:
			if _, ok := h.clients[client.user.ID]; ok {
				delete(h.clients, client.user.ID)
				close(client.sendCh)
			}
		}
	}
}
