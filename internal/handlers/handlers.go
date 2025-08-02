package handlers

import (
	"htmx/internal/components/layouts"
	"htmx/internal/components/pages"
	"htmx/internal/components/partials"
	"htmx/internal/middleware"
	"htmx/internal/models"
	"htmx/static"
	"log"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// render renders Templ components
func render(c *gin.Context, status int, template templ.Component) error {
	c.Header("Content-Type", "text/html")
	c.Status(status)
	return template.Render(c.Request.Context(), c.Writer)
}

// Hub manages WebSocket connections and broadcasts
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

// run processes WebSocket connection events
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

// upgrader configures WebSocket connection parameters
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity; restrict in production
	},
}

// WS handles WebSocket connections
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

// Handler contains dependencies for HTTP request handling
type Handler struct {
	RoomStore *models.RoomStore
	ChatStore *models.ChatStore
}

// NewHandler creates a handler with the given dependencies
func NewHandler(roomStore *models.RoomStore, chatStore *models.ChatStore) *Handler {
	return &Handler{
		RoomStore: roomStore,
		ChatStore: chatStore,
	}
}

// StartHub initializes the WebSocket hub
func StartHub() {
	go hub.run()
}

// SetupRoutes configures all application routes
func (h *Handler) SetupRoutes(router *gin.Engine) {
	// Apply middleware
	router.Use(middleware.NoCacheMiddleware())

	// Serve embedded static files
	router.StaticFS("/static/css", static.GetCSSFileSystem())
	router.StaticFS("/static/js", static.GetJSFileSystem())

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

// Home handles the home page request
func (h *Handler) Home(c *gin.Context) {
	roomID := c.Param("id")
	room, _ := h.RoomStore.GetRoom(roomID)

	homePage := pages.HomePage(room, h.ChatStore.GetChats(), h.RoomStore.GetRooms())

	if c.Request.Header.Get("HX-Request") == "true" {
		// Return just the home content for HTMX requests
		render(c, http.StatusOK, homePage)
		return
	}

	// Return full page with layout
	fullPage := layouts.Base("Chat Rooms", homePage)
	render(c, http.StatusOK, fullPage)
}

// Test handles the test page request
func (h *Handler) Test(c *gin.Context) {
	// Return full page with layout
	fullPage := layouts.Test("Chat Rooms")
	render(c, http.StatusOK, fullPage)
}

// RoomDetail handles room detail page requests
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
	roomPage := pages.HomePage(room, chats, rooms)
	fullPage := layouts.Base(room.Name, roomPage)
	render(c, http.StatusOK, fullPage)
}

// GetRooms returns the rooms list for HTMX requests
func (h *Handler) GetRooms(c *gin.Context) {
	rooms := h.RoomStore.GetRooms()
	roomsList := partials.RoomsList(rooms)
	render(c, http.StatusOK, roomsList)
}

// GetRoomsContent returns rooms list content for HTMX updates
func (h *Handler) GetRoomsContent(c *gin.Context) {
	rooms := h.RoomStore.GetRooms()
	roomsContent := partials.RoomsListContent(rooms)
	render(c, http.StatusOK, roomsContent)
}

// CreateRoom handles new room creation
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

// GetChats returns chat messages for a room
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

// CreateChat handles new chat message creation
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
