package main

import (
	"github.com/gin-gonic/gin"
	"htmx/internal/handlers"
	"htmx/internal/models"
	"log"
	"time"
)

func main() {
	// Create data stores
	roomStore := models.NewRoomStore()
	chatStore := models.NewChatStore()

	// Add some sample data
	addSampleData(roomStore, chatStore)

	// Create handler
	handler := handlers.NewHandler(roomStore, chatStore)

	// Set up Gin router
	router := gin.Default()

	// Set up routes
	handler.SetupRoutes(router)

	// Start WebSocket hub
	handlers.StartHub()

	// Start server
	log.Println("Server starting on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// addSampleData adds some sample rooms and chats for demonstration
func addSampleData(roomStore *models.RoomStore, chatStore *models.ChatStore) {
	now := time.Now()

	// Add sample rooms
	generalRoom := &models.Room{
		ID:        "1",
		Name:      "General",
		CreatedAt: now.Add(-24 * time.Hour),
	}
	techRoom := &models.Room{
		ID:        "2",
		Name:      "Technology",
		CreatedAt: now.Add(-2 * time.Hour),
	}

	roomStore.AddRoom(generalRoom)
	roomStore.AddRoom(techRoom)

	// Add sample chats
	chatStore.AddChat(&models.Chat{
		ID:        "1",
		RoomID:    "1",
		Username:  "Alice",
		Message:   "Hello everyone!",
		CreatedAt: now.Add(-20 * time.Minute),
	})

	chatStore.AddChat(&models.Chat{
		ID:        "2",
		RoomID:    "1",
		Username:  "Bob",
		Message:   "Hi Alice, how are you?",
		CreatedAt: now.Add(-15 * time.Minute),
	})

	chatStore.AddChat(&models.Chat{
		ID:        "3",
		RoomID:    "2",
		Username:  "Charlie",
		Message:   "Anyone interested in Go programming?",
		CreatedAt: now.Add(-5 * time.Minute),
	})
}
