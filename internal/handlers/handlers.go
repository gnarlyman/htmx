package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"htmx/internal/models"
	"log"
	"net/http"
	"time"
)

// WebSocket Hub for broadcasting updates
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

var hub = &Hub{
	clients:    make(map[*websocket.Conn]bool),
	broadcast:  make(chan []byte),
	register:   make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
}

func (h *Hub) run() {
	for {
		select {
		case conn := <-h.register:
			h.clients[conn] = true
		case conn := <-h.unregister:
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
		case message := <-h.broadcast:
			for conn := range h.clients {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					conn.Close()
					delete(h.clients, conn)
				}
			}
		}
	}
}

// WebSocket Upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity; restrict in production
	},
}

// WS Handler
func (h *Handler) WS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	hub.register <- conn

	go func() {
		defer func() {
			hub.unregister <- conn
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()
}

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

// StartHub starts the WebSocket hub
func StartHub() {
	go hub.run()
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
	router.GET("/api/rooms/:id/chat-content", h.GetChatContent) // New for full chat partial
	router.GET("/ws", h.WS)

	// Start hub in a goroutine
	go hub.run()
}

// Home renders the home page
func (h *Handler) Home(c *gin.Context) {
	data := gin.H{
		"title": "Chat Rooms",
		"rooms": h.RoomStore.GetRooms(),
		"Page":  "home",
	}

	if c.Request.Header.Get("HX-Request") == "true" {
		c.HTML(http.StatusOK, "partials/index-content.html", data)
		return
	}

	c.HTML(http.StatusOK, "layouts/base.html", data)
}

// RoomDetail renders the room detail page
func (h *Handler) RoomDetail(c *gin.Context) {
	roomID := c.Param("id")
	room, exists := h.RoomStore.GetRoom(roomID)
	if !exists {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	data := gin.H{
		"title": room.Name,
		"rooms": h.RoomStore.GetRooms(), // For sidebar
		"room":  room,
		"chats": h.ChatStore.GetChatsByRoom(roomID),
		"Page":  "room",
	}

	if c.Request.Header.Get("HX-Request") == "true" {
		c.HTML(http.StatusOK, "partials/chat-content.html", data)
		return
	}

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

	// Broadcast update
	hub.broadcast <- []byte("new-room")

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
		c.Status(http.StatusNotFound)
		return
	}

	var input struct {
		Username string `form:"username" binding:"required"`
		Message  string `form:"message" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
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

	// Broadcast update (could be room-specific, but global for simplicity)
	hub.broadcast <- []byte("new-chat")

	c.HTML(http.StatusOK, "partials/chats.html", gin.H{
		"chats":  h.ChatStore.GetChatsByRoom(roomID),
		"roomID": roomID,
	})
	c.Writer.Write([]byte(`<div id="chat-form-error" hx-swap-oob="innerHTML"></div>`))
}

// GetChatContent returns the full chat content partial for HTMX swaps
func (h *Handler) GetChatContent(c *gin.Context) {
	roomID := c.Param("id")
	room, exists := h.RoomStore.GetRoom(roomID)
	if !exists {
		c.Status(http.StatusNotFound)
		return
	}

	data := gin.H{
		"room":  room,
		"chats": h.ChatStore.GetChatsByRoom(roomID),
	}

	c.HTML(http.StatusOK, "partials/chat-content.html", data)
}
