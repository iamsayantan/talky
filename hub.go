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

// CreateOrJoinRoom either creates a room if it does not exist in the hub and then adds the
// user to the room. If room already exists, then it just adds the user to the room.
func (h *Hub) CreateOrJoinRoom(payload CreateOrJoinRoomMessage, user *User) {
	room, ok := h.rooms[payload.RoomID]
	if !ok {
		room = NewRoom(payload.RoomType, payload.RoomID)
	}

	err := room.AddMember(user)
	if err != nil {
		log.Printf("Error while adding user to room: %v", err)
	}
	log.Printf("Added User %s to room id %s", user.Username, room.ID)
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
		case broadcastMessage := <-h.broadcastCh:
			var msg *Message
			if err := json.Unmarshal(broadcastMessage.Payload, &msg); err != nil {
				log.Printf("Error unmarshalling websocket message: %v", err)
			}

			switch msg.Type {
			case CreateOrJoinRoom:
				var payload CreateOrJoinRoomMessage
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					log.Printf("Error unmarshalling websocket payload: %v", err)
				}

				h.CreateOrJoinRoom(payload, broadcastMessage.User)
			}
		}
	}
}
