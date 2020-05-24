package talky

import (
	"errors"
	"log"
	"sync"
)

var (
	ErrAlreadyInRoom    = errors.New("already a member of the room")
	ErrRoomCapacityFull = errors.New("room capacity is full")
)

type RoomType string

const (
	AudioRoom      RoomType = "AUDIO"
	AudioVideoRoom RoomType = "AUDIO_VIDEO"
)

const (
	MaxMembersInAudioRoom      = 20
	MaxMembersInAudioVideoRoom = 4
)

// Room defines the data structure for the room where the actual call will take place.
// Room data is not store in the database, instead this will be an in memory collection
// inside the running application.
type Room struct {
	ID       string         `json:"id"`        // ID UUID string generated by client side
	RoomType RoomType       `json:"room_type"` // RoomType What kind of communication we allow inside the room is determined by this.
	Members  map[uint]*User `json:"members"`   // Members All the users who joined the room.

	mu sync.Mutex
}

// NewRoom
func NewRoom(roomType RoomType, roomId string) *Room {
	return &Room{
		ID:       roomId,
		RoomType: roomType,
		Members:  make(map[uint]*User),
	}
}

// AddMember Adds a new user to a room. Rooms have different capacity for members based on the room type.
func (r *Room) AddMember(user *User) error {
	var maxAllowedMembers int
	if r.RoomType == AudioRoom {
		maxAllowedMembers = MaxMembersInAudioRoom
	} else {
		maxAllowedMembers = MaxMembersInAudioVideoRoom
	}

	if len(r.Members) >= maxAllowedMembers {
		return ErrRoomCapacityFull
	}

	if _, ok := r.Members[user.ID]; ok {
		return ErrAlreadyInRoom
	}

	r.mu.Lock()
	r.Members[user.ID] = user
	r.mu.Unlock()

	log.Printf("Added user %s to room %s. Current members: %d", user.Username, r.ID, len(r.Members))
	return nil
}

// RemoveMember Removes the user from the rooms member list.
func (r *Room) RemoveMember(user *User) error {
	if _, ok := r.Members[user.ID]; !ok {
		// Ignoring if user who is being removed does not exist in the room.
		return nil
	}

	r.mu.Lock()
	delete(r.Members, user.ID)
	r.mu.Unlock()

	log.Printf("Removed user %s from room %s. Current members: %d", user.Username, r.ID, len(r.Members))
	return nil
}
