package ws

//REFACTOR THE ERROR DISPLAY

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/keykibatyr/triad-chat/internal/middleware"
	"github.com/keykibatyr/triad-chat/internal/repository"
	"github.com/keykibatyr/triad-chat/internal/utils"
)

type Handler struct {
	hub *Hub

	Templates struct {
		Chat utils.Template
	}

	userService repository.UserService
}

func NewHubHandler(h *Hub, userService repository.UserService) *Handler {
	return &Handler{
		hub: h,
		userService: userService,
	}
	
}

type CeateRoomReq struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
		log.Print("could not read the context")
		return
	}

	user, err := h.userService.GetByID(c, userID)
	if err != nil{
		log.Print("could not get the user from userID")
		return
	}

	clientID := strconv.FormatInt(userID, 10)

	client, err := h.hub.GetClient(clientID)
	if err == nil  {
		log.Println("connection already exists")
		log.Println(client)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}


	client = &Client{
		Hub: h.hub,
		Conn:     conn,
		Rooms:    make(map[string]*Room),
		Send:     make(chan []byte, 10),
		ID:       clientID,
		Username: user.Username,
	}

	log.Print(client)


	h.hub.AddClient(client)

	go client.writePump()
	go client.readPump()
}

func  (h *Handler) ChatPage(c *gin.Context) {
	data := gin.H{}
	utils.Render(c, h.Templates.Chat, data)
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var roomForm CeateRoomReq

	err := c.ShouldBind(&roomForm)
	if err != nil {
		c.JSON(http.StatusBadRequest, "could not read the request form")
		return
	}

	h.hub.Rooms[roomForm.ID] = h.hub.GetOrCreateRoom(roomForm.ID, roomForm.Name)

	c.JSON(http.StatusOK, roomForm)
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
