package handlers

import (
	"htmx/internal/components/layouts"
	"htmx/internal/components/pages"
	"htmx/internal/components/partials"
	"htmx/internal/models"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Helper function to render Templ components
func render(c *gin.Context, status int, template templ.Component) error {
	c.Header("Content-Type", "text/html")
	c.Status(status)
	return template.Render(c.Request.Context(), c.Writer)
}

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

// Update SetupRoutes to include the new endpoint
func (h *Handler) SetupRoutes(router *gin.Engine) {
	// Disable cache for static files in debug mode to ensure changes (e.g., CSS) are picked up immediately
	router.Use(func(c *gin.Context) {
		if gin.Mode() == gin.DebugMode && strings.HasPrefix(c.Request.URL.Path, "/static/") {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	})

	// Serve static files
	router.Static("/static", "./static")

	// HTML routes
	router.GET("/", h.Home)
	router.GET("/test", h.Test)
	router.GET("/rooms/:id", h.RoomDetail)

	// API routes for HTMX
	router.GET("/api/rooms", h.GetRooms)
	router.GET("/api/rooms-content", h.GetRoomsContent) // Add this line
	router.POST("/api/rooms", h.CreateRoom)
	router.GET("/api/rooms/:id/chats", h.GetChats)
	router.POST("/api/rooms/:id/chats", h.CreateChat)
	router.GET("/ws", h.WS)

	// Start hub in a goroutine
	go hub.run()
}

// Home renders the home page
func (h *Handler) Home(c *gin.Context) {
	rooms := h.RoomStore.GetRooms()
	homePage := pages.HomePage(rooms)

	if c.Request.Header.Get("HX-Request") == "true" {
		// Return just the home content for HTMX requests
		render(c, http.StatusOK, homePage)
		return
	}

	// Return full page with layout
	fullPage := layouts.Base("Chat Rooms", homePage)
	render(c, http.StatusOK, fullPage)
}

// Test renders the test page
func (h *Handler) Test(c *gin.Context) {
	// Return full page with layout
	fullPage := layouts.Test("Chat Rooms")
	render(c, http.StatusOK, fullPage)
}

// RoomDetail renders the room detail page OR just room content for HTMX
func (h *Handler) RoomDetail(c *gin.Context) {
	roomID := c.Param("id")
	room, exists := h.RoomStore.GetRoom(roomID)
	if !exists {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	rooms := h.RoomStore.GetRooms()
	chats := h.ChatStore.GetChatsByRoom(roomID)

	if c.Request.Header.Get("HX-Request") == "true" {
		// Return just the room content for HTMX requests
		roomContent := partials.RoomContent(room, chats)
		render(c, http.StatusOK, roomContent)
		return
	}

	// Return full page with layout
	roomPage := pages.RoomPage(room, chats, rooms)
	fullPage := layouts.Base(room.Name, roomPage)
	render(c, http.StatusOK, fullPage)
}

// GetRooms returns the rooms list partial for HTMX
func (h *Handler) GetRooms(c *gin.Context) {
	rooms := h.RoomStore.GetRooms()
	roomsList := partials.RoomsList(rooms)
	render(c, http.StatusOK, roomsList)
}

// GetRoomsContent returns just the rooms list content for HTMX updates
func (h *Handler) GetRoomsContent(c *gin.Context) {
	rooms := h.RoomStore.GetRooms()
	roomsContent := partials.RoomsListContent(rooms)
	render(c, http.StatusOK, roomsContent)
}

// CreateRoom creates a new room
func (h *Handler) CreateRoom(c *gin.Context) {
	var input struct {
		Name string `form:"name" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString(`<div class="text-error">Room name is required</div>`)
		return
	}

	room := &models.Room{
		ID:        uuid.New().String(),
		Name:      input.Name,
		CreatedAt: time.Now(),
	}

	h.RoomStore.AddRoom(room)

	// Broadcast to other users
	go func() {
		hub.broadcast <- []byte("new-room")
	}()

	// Return ONLY the rooms content (not the full component with form)
	rooms := h.RoomStore.GetRooms()
	roomsContent := partials.RoomsListContent(rooms)
	render(c, http.StatusOK, roomsContent)
}

// GetChats returns the chats list partial for HTMX
func (h *Handler) GetChats(c *gin.Context) {
	roomID := c.Param("id")
	_, exists := h.RoomStore.GetRoom(roomID)
	if !exists {
		c.Status(http.StatusNotFound)
		return
	}

	chats := h.ChatStore.GetChatsByRoom(roomID)
	messagesList := partials.MessagesList(chats)
	render(c, http.StatusOK, messagesList)
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
		c.Header("Content-Type", "text/html")
		c.Status(http.StatusBadRequest)
		c.Writer.WriteString(`<div class="text-error">Username and message are required</div>`)
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

	// Broadcast update
	hub.broadcast <- []byte("new-chat")

	// Return updated messages list
	chats := h.ChatStore.GetChatsByRoom(roomID)
	messagesList := partials.MessagesList(chats)
	render(c, http.StatusOK, messagesList)

	// Clear error message
	c.Writer.Write([]byte(`<div id="chat-form-error" hx-swap-oob="innerHTML"></div>`))
}
