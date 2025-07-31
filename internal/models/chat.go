package models

import (
	"sync"
	"time"
)

// Chat represents a chat message in a room
type Chat struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatStore manages the collection of chats
type ChatStore struct {
	chats map[string]*Chat
	// Secondary index by room ID for quick access
	chatsByRoom map[string][]*Chat
	mutex       sync.RWMutex
}

// NewChatStore creates a new chat store
func NewChatStore() *ChatStore {
	return &ChatStore{
		chats:       make(map[string]*Chat),
		chatsByRoom: make(map[string][]*Chat),
	}
}

// GetChats returns all chats
func (s *ChatStore) GetChats() []*Chat {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	chats := make([]*Chat, 0, len(s.chats))
	for _, chat := range s.chats {
		chats = append(chats, chat)
	}
	return chats
}

// GetChat returns a chat by ID
func (s *ChatStore) GetChat(id string) (*Chat, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	chat, exists := s.chats[id]
	return chat, exists
}

// GetChatsByRoom returns all chats for a specific room
func (s *ChatStore) GetChatsByRoom(roomID string) []*Chat {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy to prevent concurrent modification issues
	chats := make([]*Chat, len(s.chatsByRoom[roomID]))
	copy(chats, s.chatsByRoom[roomID])
	return chats
}

// AddChat adds a new chat message
func (s *ChatStore) AddChat(chat *Chat) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.chats[chat.ID] = chat
	s.chatsByRoom[chat.RoomID] = append(s.chatsByRoom[chat.RoomID], chat)
}

// DeleteChat removes a chat message
func (s *ChatStore) DeleteChat(id string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	chat, exists := s.chats[id]
	if !exists {
		return false
	}

	// Remove from main map
	delete(s.chats, id)

	// Remove from room index
	roomChats := s.chatsByRoom[chat.RoomID]
	for i, c := range roomChats {
		if c.ID == id {
			// Remove this chat from the slice
			s.chatsByRoom[chat.RoomID] = append(roomChats[:i], roomChats[i+1:]...)
			break
		}
	}

	return true
}

// DeleteChatsByRoom removes all chats for a specific room
func (s *ChatStore) DeleteChatsByRoom(roomID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Get all chats for this room
	roomChats := s.chatsByRoom[roomID]

	// Remove each chat from the main map
	for _, chat := range roomChats {
		delete(s.chats, chat.ID)
	}

	// Clear the room index
	delete(s.chatsByRoom, roomID)
}
