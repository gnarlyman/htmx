package models

import (
	"sync"
	"time"
)

// Room represents a chat room
type Room struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// RoomStore manages the collection of rooms
type RoomStore struct {
	rooms map[string]*Room
	mutex sync.RWMutex
}

// NewRoomStore creates a new room store
func NewRoomStore() *RoomStore {
	return &RoomStore{
		rooms: make(map[string]*Room),
	}
}

// GetRooms returns all rooms
func (s *RoomStore) GetRooms() []*Room {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	rooms := make([]*Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// GetRoom returns a room by ID
func (s *RoomStore) GetRoom(id string) (*Room, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	room, exists := s.rooms[id]
	return room, exists
}

// AddRoom adds a new room
func (s *RoomStore) AddRoom(room *Room) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.rooms[room.ID] = room
}

// UpdateRoom updates an existing room
func (s *RoomStore) UpdateRoom(room *Room) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.rooms[room.ID]; !exists {
		return false
	}

	s.rooms[room.ID] = room
	return true
}

// DeleteRoom removes a room
func (s *RoomStore) DeleteRoom(id string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.rooms[id]; !exists {
		return false
	}

	delete(s.rooms, id)
	return true
}
