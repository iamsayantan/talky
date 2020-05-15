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
	hub := &Hub{
		rooms:        make(map[string]*Room),
		clients:      make(map[uint]*Client),
		registerCh:   make(chan *Client),
		unregisterCh: make(chan *Client),
		broadcastCh:  make(chan *BroadcastMessage),
	}

	go hub.run()
	return hub
}

func (h *Hub) AddClient(client *Client) {
	h.registerCh <- client
}

func (h *Hub) RemoveClient(client *Client) {
	h.unregisterCh <- client
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.registerCh:
			log.Printf("Registering new client with user id: %d", client.user.ID)
			h.clients[client.user.ID] = client
		case client := <-h.unregisterCh:
			log.Printf("Removing client with user id: %d", client.user.ID)
			if _, ok := h.clients[client.user.ID]; ok {
				delete(h.clients, client.user.ID)
				close(client.sendCh)
			}
		}
	}
}
