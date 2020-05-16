package talky

import (
	"encoding/json"
	"log"
)

type Hub struct {
	rooms   map[string]*Room
	clients map[uint]*Client

	registerCh   chan *Client
	unregisterCh chan *Client
	broadcastCh  chan []byte
}

func NewHub() *Hub {
	hub := &Hub{
		rooms:        make(map[string]*Room),
		clients:      make(map[uint]*Client),
		registerCh:   make(chan *Client),
		unregisterCh: make(chan *Client),
		broadcastCh:  make(chan []byte),
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
		case message := <-h.broadcastCh:
			log.Printf("Iside hub broadcast")
			var msg *Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error unmarshalling websocket message: %v", err)
			}

			switch msg.Type {
			case CreateOrJoinRoom:
				var payload CreateOrJoinRoomMessage
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					log.Printf("Error unmarshalling websocket payload: %v", err)
				}

				log.Printf("[Websocket] Message Type: %s, Message Payload: %v", msg.Type, payload.RoomType)
			}
		}
	}
}
