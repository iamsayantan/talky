package talky

import "encoding/json"

const (
	CreateOrJoinRoom = "CREATE_OR_JOIN"
	Offer            = "OFFER"
	Answer           = "ANSWER"
	ICECandidate     = "ICE_CANDIDATE"
	RoomJoin         = "ROOM_JOIN"
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

// CreateOrJoinRoomMessage is the payload sent when someone creates or joins a new room.
type CreateOrJoinRoomMessage struct {
	RoomID   string   `json:"room_id"`
	RoomType RoomType `json:"room_type"`
}

type RoomMessage struct {
	RoomID       string `json:"room_id"`        // RoomID is id of the room for where the SDPMessage is intended.
	UserID       uint   `json:"user_id"`        // UserID is the user who sent the message. The message will not be sent to this user.
	TargetUserID uint   `json:"target_user_id"` // TargetUserID holds the id of the user if the message is sent specifically to this user.
}

// SDPMessage is the payload for session descriptions in a room.
type SDPMessage struct {
	RoomMessage
	SDP interface{} `json:"sdp"` // SDP is the actual SDP payload.
}

type ICEMessage struct {
	RoomMessage
	Candidate interface{} `json:"candidate"`
}

type ResponseMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type RoomJoined struct {
	RoomID      string `json:"room_id"`
	User        User   `json:"user"`
	IsInitiator bool   `json:"is_initiator"`
}
