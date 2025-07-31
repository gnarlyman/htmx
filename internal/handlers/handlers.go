package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"htmx/internal/models"
	"log"
	"net/http"
	"time"
)

// Handler holds the dependencies for all handlers
type Handler struct {
	RoomStore *models.RoomStore
	ChatStore *models.ChatStore
}

// NewHandler creates a new handler with the given dependencies
func NewHandler(roomStore *models.RoomStore, chatStore *models.ChatStore) *Handler {
	return &Handler{
		RoomStore: roomStore,
		ChatStore: chatStore,
	}
}

// SetupRoutes configures all the routes for our application
func (h *Handler) SetupRoutes(router *gin.Engine) {
	// Serve static files
	router.Static("/static", "./static")

	// HTML routes
	router.GET("/", h.Home)
	router.GET("/rooms/:id", h.RoomDetail)

	// API routes for HTMX
	router.GET("/api/rooms", h.GetRooms)
	router.POST("/api/rooms", h.CreateRoom)
	router.GET("/api/rooms/:id/chats", h.GetChats)
	router.POST("/api/rooms/:id/chats", h.CreateChat)
}

// Home renders the home page
func (h *Handler) Home(c *gin.Context) {
	data := gin.H{
		"title": "Chat Rooms",
		"rooms": h.RoomStore.GetRooms(),
	}

	if c.Request.Header.Get("HX-Request") == "true" {
		c.HTML(http.StatusOK, "partials/index-content.html", data)
		return
	}

	data["Page"] = "home"
	c.HTML(http.StatusOK, "layouts/base.html", data)
}

// RoomDetail renders the room detail page
func (h *Handler) RoomDetail(c *gin.Context) {
	roomID := c.Param("id")
	room, exists := h.RoomStore.GetRoom(roomID)
	if !exists {
		log.Printf("Room %s not found, redirecting to /", roomID)
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	data := gin.H{
		"title": room.Name,
		"room":  room,
		"chats": h.ChatStore.GetChatsByRoom(roomID),
	}

	if c.Request.Header.Get("HX-Request") == "true" {
		c.HTML(http.StatusOK, "partials/room-content.html", data)
		return
	}

	data["Page"] = "room"
	c.HTML(http.StatusOK, "layouts/base.html", data)
}

// GetRooms returns the rooms list partial for HTMX
func (h *Handler) GetRooms(c *gin.Context) {
	c.HTML(http.StatusOK, "partials/rooms.html", gin.H{
		"rooms": h.RoomStore.GetRooms(),
	})
}

// CreateRoom creates a new room
func (h *Handler) CreateRoom(c *gin.Context) {
	var input struct {
		Name string `form:"name" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		log.Printf("CreateRoom error: %v", err)
		c.HTML(http.StatusBadRequest, "partials/room-form.html", gin.H{
			"error": "Room name is required",
		})
		return
	}

	room := &models.Room{
		ID:        uuid.New().String(),
		Name:      input.Name,
		CreatedAt: time.Now(),
	}

	h.RoomStore.AddRoom(room)

	// Return the updated rooms list and clear error
	c.HTML(http.StatusOK, "partials/rooms.html", gin.H{
		"rooms": h.RoomStore.GetRooms(),
	})
	c.Writer.Write([]byte(`<div id="room-form-error" hx-swap-oob="innerHTML"></div>`))
}

// GetChats returns the chats list partial for HTMX
func (h *Handler) GetChats(c *gin.Context) {
	roomID := c.Param("id")
	_, exists := h.RoomStore.GetRoom(roomID)
	if !exists {
		log.Printf("Room %s not found for GetChats", roomID)
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "partials/chats.html", gin.H{
		"chats":  h.ChatStore.GetChatsByRoom(roomID),
		"roomID": roomID,
	})
}

// CreateChat creates a new chat message
func (h *Handler) CreateChat(c *gin.Context) {
	roomID := c.Param("id")
	_, exists := h.RoomStore.GetRoom(roomID)
	if !exists {
		log.Printf("Room %s not found for CreateChat", roomID)
		c.Status(http.StatusNotFound)
		return
	}

	var input struct {
		Username string `form:"username" binding:"required"`
		Message  string `form:"message" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		log.Printf("CreateChat error: %v", err)
		c.HTML(http.StatusBadRequest, "partials/chat-form.html", gin.H{
			"error":  "Username and message are required",
			"roomID": roomID,
		})
		return
	}

	chat := &models.Chat{
		ID:        uuid.New().String(),
		RoomID:    roomID,
		Username:  input.Username,
		Message:   input.Message,
		CreatedAt: time.Now(),
	}

	h.ChatStore.AddChat(chat)

	// Return the updated chats list and clear error
	c.HTML(http.StatusOK, "partials/chats.html", gin.H{
		"chats":  h.ChatStore.GetChatsByRoom(roomID),
		"roomID": roomID,
	})
	c.Writer.Write([]byte(`<div id="chat-form-error" hx-swap-oob="innerHTML"></div>`))
}
