package talky

import "encoding/json"

const (
	CreateOrJoinRoom = "CREATE_OR_JOIN"
)

// Message type represents the basic message type received from client.
// Depending on the type, the payload structure can be different.
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type CreateOrJoinRoomMessage struct {
	RoomID   string `json:"room_id"`
	RoomType string `json:"room_type"`
}
