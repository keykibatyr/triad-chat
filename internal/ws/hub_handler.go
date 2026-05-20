package ws

//REFACTOR THE ERROR DISPLAY

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/keykibatyr/triad-chat/internal/chat/service"
	"github.com/keykibatyr/triad-chat/internal/middleware"
	"github.com/keykibatyr/triad-chat/internal/repository"
	"github.com/keykibatyr/triad-chat/internal/utils"
)

const messageLimit = 30

type Handler struct {
	hub       *Hub
	ctx       context.Context
	Templates struct {
		Chat utils.Template
	}

	userService    repository.UserService
	messageService service.MessageServiceInterface
}

func NewHubHandler(h *Hub, userService repository.UserService, messageService service.MessageServiceInterface, ctx context.Context) *Handler {
	return &Handler{
		hub:            h,
		ctx:            ctx,
		userService:    userService,
		messageService: messageService,
	}

}

type CeateRoomReq struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type SearchRoomReq struct {
	Name string `json:"name"`
}

type LoadMessagesReq struct {
	ConvoID string `json:"convoID"`
}

type LoadMessagesAndCursorReq struct {
	ConvoID string `json:"convoID"`
	Cursor  string `json:"cursor"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) ServeWs(c *gin.Context) {

	userID, _, err := middleware.CurrentUser(c)
	if err != nil {
		return
	}

	user, err := h.userService.GetByID(c, userID)
	if err != nil {
		return
	}


	clientID := strconv.FormatInt(userID, 10)


	client, err := h.hub.GetClient(clientID)
	if err == nil {
		client.Conn.Close()
	}


	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}


	client = &Client{
		Hub:      h.hub,
		Conn:     conn,
		Rooms:    make(map[string]*Room),
		Send:     make(chan []byte, 10),
		ID:       clientID,
		Username: user.Username,
	}

	log.Print(client)

	h.hub.AddClient(client)

	go client.writePump()
	go client.readPump(h.ctx)

}

func (h *Handler) ChatPage(c *gin.Context) {
	data := gin.H{}
	utils.Render(c, h.Templates.Chat, data)
}

func (h *Handler) ShowBySearch(c *gin.Context) {
	var SearchForm SearchRoomReq

	err := c.ShouldBindJSON(&SearchForm)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid search form"})
		return
	}

	convos, err := h.hub.GetRooms(c, SearchForm.Name)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get roooms"})
		return
	}

	c.JSON(200, convos)

}

func (h *Handler) CreateRoom(c *gin.Context) {
	var roomForm CeateRoomReq

	err := c.ShouldBindJSON(&roomForm)
	if err != nil {
		c.JSON(http.StatusBadRequest, "could not read the request form")
		return
	}
	h.hub.CreateRoom(c, roomForm.Name, roomForm.Type)

	c.JSON(http.StatusOK, roomForm)
}

func (h *Handler) ChatLoad(c *gin.Context) {
	convoIDStr := c.Query("convoID")
	cursorStr := c.Query("cursor")
	convoID, _ := strconv.ParseInt(convoIDStr, 10, 64)

	if cursorStr == "" {
		paginate, err := h.messageService.ListMessagesAtLoad(c, convoID, messageLimit)
		if err != nil {
			c.JSON(http.StatusBadRequest, "could not load message history")
			return
		}

		c.JSON(http.StatusOK, paginate)
		return
	}

	cursor, _ := strconv.Atoi(cursorStr)
	paginate, err := h.messageService.ListMessagesAndCursor(c, convoID, messageLimit, cursor)
	if err != nil {
		c.JSON(http.StatusBadRequest, "could not load message history with cursor")
		return
	}

	c.JSON(http.StatusOK, paginate)

}

// func (h *Handler) JoinRoom(c *gin.Context) {
// 	roomID := c.Param("roomID")

// 	userID, _, err := middleware.CurrentUser(c)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, "could not read the context")
// 		return
// 	}

// 	clientID := strconv.FormatInt(userID, 10)

// 	client, err := h.hub.GetClient(clientID)
// 	if err != nil {
// 		log.Print("clinet doesnt exist")
// 		return
// 	}

// 	message := &Message{
// 		Type: "join_room",
// 		Content:  "a new user has joined the chat",
// 		Username: client.Username,
// 		RoomID:   roomID,
// 	}

// 	room, ok := h.hub.Rooms[roomID]
// 	if !ok {
// 		c.JSON(http.StatusInternalServerError, "could not get the room from hub")
// 		return
// 	}

// 	client.Rooms[roomID] = room

// 	client.JoinRoom(room)
// 	room.Broadcast <- message

// 	// func (h *Hub) AddClient(client *Client) {
// 	// h.mu.Lock()
// 	// defer h.mu.Unlock()
// 	// h.Clients[client.ID] = client
// 	// }
// }

// func (h *Handler) LeaveRoom(c *gin.Context) {

// 	userID, _, err := middleware.CurrentUser(c)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, "could not read the context")
// 		return
// 	}

// 	roomID := c.Param("roomID")
// 	// roomName := c.Query("name")
// 	clientID := strconv.FormatInt(userID, 10)

// 	client, err := h.hub.GetClient(clientID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, "could not get the client from hub")
// 		return
// 	}

// 	message := &Message{
// 		Type: "leave_room",
// 		Content:  "a new user has left the chat",
// 		Username: client.Username,
// 		RoomID:   roomID,
// 	}

// 	room, ok := client.Rooms[roomID]
// 	if !ok {
// 		c.JSON(http.StatusInternalServerError, "could not get the room from hub")
// 		return
// 	}

// 	client.LeaveRoom(room)

// 	if client.Rooms == nil {
// 		h.hub.DeleteClient(client)
// 	}

// 	// func (h *Hub) DeleteClient(client *Client) {
// 	// 	h.mu.Lock()
// 	// 	defer h.mu.Unlock()
// 	// 	delete(h.Clients, client.ID)
// 	// }

// 	room.Broadcast <- message
// }
