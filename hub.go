package talky

import (
	"encoding/json"
	"errors"
	"log"
)

type Hub struct {
	rooms       map[string]*Room
	clients     map[uint]*Client
	clientRooms map[uint]*Room // a single user can only be part of one room, so the hub keeps that mapping

	registerCh   chan *Client
	unregisterCh chan *Client
	broadcastCh  chan *BroadcastMessage
}

func NewHub() *Hub {
	hub := &Hub{
		rooms:        make(map[string]*Room),
		clients:      make(map[uint]*Client),
		clientRooms:  make(map[uint]*Room),
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
	h.RoomCleanup(client)
	h.unregisterCh <- client
}

// RoomCleanup removes the user from any room he is part of. Also removes the room
// from the hub if it becomes empty.
func (h *Hub) RoomCleanup(client *Client) {
	room, ok := h.clientRooms[client.user.ID]
	if !ok {
		return
	}

	_ = room.RemoveMember(client.user)
	delete(h.clientRooms, client.user.ID)

	// remove empty rooms from the memory.
	if len(room.Members) == 0 {
		log.Printf("Room %s is empty, removing from the Hub", room.ID)
		delete(h.rooms, room.ID)
	}
}

// CreateOrJoinRoom either creates a room if it does not exist in the hub and then adds the
// user to the room. If room already exists, then it just adds the user to the room.
func (h *Hub) CreateOrJoinRoom(payload CreateOrJoinRoomMessage, user *User) error {
	// isInitiator is used to track if the room is initiated by the user, if a room is not available in the
	// room list in hub, then we assume the first user a initiator.
	isInitiator := false
	room, ok := h.rooms[payload.RoomID]
	if !ok {
		isInitiator = true
		room = NewRoom(payload.RoomType, payload.RoomID)
		h.rooms[room.ID] = room
	}

	// at a time a single user can be part of only one room.
	if existingRoom, ok := h.clientRooms[user.ID]; ok && existingRoom.ID != room.ID {
		return errors.New("you are already a part of a room")
	}

	err := room.AddMember(user)
	if err != nil {
		log.Printf("Error while adding user to room: %v", err)
		return err
	}

	h.clientRooms[user.ID] = room

	roomJoined := RoomJoined{
		RoomID:      room.ID,
		User:        *user,
		IsInitiator: isInitiator,
	}

	responsePayload := ResponseMessage{
		Type:    RoomJoin,
		Payload: roomJoined,
	}

	resp, _ := json.Marshal(responsePayload)

	// RoomJoin message should be broadcast to all users in the room.
	for _, member := range room.Members {
		if client, ok := h.clients[member.ID]; ok {
			client.sendCh <- resp
		}
	}

	return nil
}

func (h *Hub) HandleHangup(payload HangupCall, user *User) error {
	room, ok := h.clientRooms[payload.UserID]
	if !ok {
		return nil
	}

	err := room.RemoveMember(user)
	if err != nil {
		log.Printf("Error while removing user from room: %v", err)
		return err
	}
	delete(h.clientRooms, user.ID)

	responsePayload := ResponseMessage{
		Type:    Hangup,
		Payload: payload,
	}

	resp, _ := json.Marshal(responsePayload)
	for _, member := range room.Members {
		if client, ok := h.clients[member.ID]; ok {
			if client.user.ID == user.ID {
				continue
			}
			client.sendCh <- resp
		}
	}

	return nil
}

func (h *Hub) PropagateSDPOffer(payload SDPMessage) error {
	room, ok := h.rooms[payload.RoomID]
	if !ok {
		return errors.New("room not found")
	}

	responsePayload := ResponseMessage{
		Type:    Offer,
		Payload: payload,
	}

	resp, _ := json.Marshal(responsePayload)

	for _, member := range room.Members {
		if client, ok := h.clients[member.ID]; ok {
			// we exclude the the user who sent the offer.
			if client.user.ID == payload.User.ID {
				continue
			}

			client.sendCh <- resp
		}
	}

	return nil
}

func (h *Hub) SendAnswer(payload SDPMessage) error {
	room, ok := h.rooms[payload.RoomID]
	if !ok {
		return errors.New("room not found")
	}

	responsePayload := ResponseMessage{
		Type:    Answer,
		Payload: payload,
	}

	resp, _ := json.Marshal(responsePayload)
	for _, member := range room.Members {
		if client, ok := h.clients[member.ID]; ok {
			// answers should only be sent to the targeted user.
			if payload.TargetUserID == 0 || client.user.ID != payload.TargetUserID {
				log.Printf("Skipping sending answer to %d", client.user.ID)
				continue
			}

			log.Printf("Sending answer to %d", client.user.ID)
			client.sendCh <- resp
		}
	}

	return nil
}

func (h *Hub) SendICE(payload ICEMessage) error {
	room, ok := h.rooms[payload.RoomID]
	if !ok {
		return errors.New("room not found")
	}

	responsePayload := ResponseMessage{
		Type:    ICECandidate,
		Payload: payload,
	}
	resp, _ := json.Marshal(responsePayload)
	for _, member := range room.Members {
		if client, ok := h.clients[member.ID]; ok {
			if client.user.ID == payload.User.ID {
				continue
			}

			client.sendCh <- resp
		}
	}

	return nil
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

				err := h.CreateOrJoinRoom(payload, broadcastMessage.User)
				if err != nil {
					errPayload := ResponseMessage{
						Type:    "error",
						Payload: err.Error(),
					}

					msg, _ := json.Marshal(errPayload)
					h.clients[broadcastMessage.User.ID].sendCh <- msg
				}
			case Hangup:
				var payload HangupCall
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					log.Printf("Error unmarshalling websocket paylaod: %v", err)
				}

				err := h.HandleHangup(payload, broadcastMessage.User)
				if err != nil {
					errPayload := ResponseMessage{
						Type:    "error",
						Payload: err.Error(),
					}

					msg, _ := json.Marshal(errPayload)
					h.clients[broadcastMessage.User.ID].sendCh <- msg
				}
			case Offer:
				var payload SDPMessage
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					log.Printf("Error unmarshalling websocket paylaod: %v", err)
				}

				err := h.PropagateSDPOffer(payload)
				if err != nil {
					errPayload := ResponseMessage{
						Type:    "error",
						Payload: err.Error(),
					}

					msg, _ := json.Marshal(errPayload)
					h.clients[broadcastMessage.User.ID].sendCh <- msg
				}
			case Answer:
				var payload SDPMessage
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					log.Printf("Error unmarshalling websocket paylaod: %v", err)
				}

				err := h.SendAnswer(payload)
				if err != nil {
					errPayload := ResponseMessage{
						Type:    "error",
						Payload: err.Error(),
					}

					msg, _ := json.Marshal(errPayload)
					h.clients[broadcastMessage.User.ID].sendCh <- msg
				}
			case ICECandidate:
				var payload ICEMessage
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					log.Printf("Error unmarshalling websocket paylaod: %v", err)
				}

				err := h.SendICE(payload)
				if err != nil {
					errPayload := ResponseMessage{
						Type:    "error",
						Payload: err.Error(),
					}

					msg, _ := json.Marshal(errPayload)
					h.clients[broadcastMessage.User.ID].sendCh <- msg
				}
			}
		}
	}
}
