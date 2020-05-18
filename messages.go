package talky

import "encoding/json"

const (
	CreateOrJoinRoom = "CREATE_OR_JOIN"
)

// BroadcastMessage defines the type for broadcast message.
type BroadcastMessage struct {
	User    *User  // User from whom we got the message
	Payload []byte // Payload the message
}

// Message type represents the basic message type exchanged with clients.
// Depending on the type, the payload structure can be different.
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type CreateOrJoinRoomMessage struct {
	RoomID   string   `json:"room_id"`
	RoomType RoomType `json:"room_type"`
}

type ResponseMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
