package ws

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/keykibatyr/triad-chat/internal/chat/models"
	"github.com/keykibatyr/triad-chat/internal/chat/service"
	"github.com/keykibatyr/triad-chat/internal/chat/worker"
)

type Hub struct {
	Rooms   map[string]*Room
	Clients map[string]*Client
	mu      sync.RWMutex

	Worker       *worker.PersistWorker
	ConvoService service.ConversationServiceInterface
	MessageService service.MessageServiceInterface
	ConvoSummaryService service.ConversationSummaryServiceInterface
}

func NewHub(Worker *worker.PersistWorker, 
	convoService service.ConversationServiceInterface, 
	messageService service.MessageServiceInterface,
	convoSummaryService service.ConversationSummaryServiceInterface,
	) *Hub {
	return &Hub{
		Rooms:   make(map[string]*Room),
		Clients: make(map[string]*Client),

		Worker:       Worker,
		ConvoService: convoService,
		MessageService: messageService,
		ConvoSummaryService: convoSummaryService,
	}
}

func (h *Hub) CreateRoom(ctx context.Context, name, convoType string) *Room {
	convo, err := h.ConvoService.SaveConvoToDB(ctx, name, convoType)
	if err != nil {
		log.Println("unable to save convo to db")
		return nil
	}

	id := strconv.FormatInt(convo.ID, 10)
	convosummary, err := h.ConvoSummaryService.InitialSummary(ctx, convo.ID)
	if err != nil {
		log.Println("unable to create an i")
		return nil
	}

	h.mu.Lock()
	room := NewRoom(id, convo.Name)
	h.Rooms[id] = room
	room.RoomSummary = convosummary.SummaryText
	h.mu.Unlock()
	go room.run()
	return room
}

func (h *Hub) GetRooms(ctx context.Context, name string) ([]models.Convo, error) {
	convoList, err := h.ConvoService.GetConvosFromDB(ctx, name)
	
	if err != nil {
		return nil, fmt.Errorf("could not get convos from DB: %w", err)
	}
	// h.mu.RLock()
	// defer h.mu.RUnlock()
	
	return convoList, nil
}

func (h *Hub) GetOrCreateRoom(ctx context.Context, id string) *Room {
	h.mu.RLock()
	room, ok := h.Rooms[id]
	h.mu.RUnlock()
	if ok {
		return room
	}

	id64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Panic("could not convert string to in64")
		return nil
	}
	log.Println(id64)

	convo, err := h.ConvoService.GetConvoByID(ctx, id64)
	if err != nil {
		log.Panic("bug err")
		log.Println(convo)
		return nil
	}
	log.Println(convo)

	h.mu.Lock()
	room, ok = h.Rooms[id]
	h.mu.Unlock()
	if ok {
		return room
	}
	
	h.mu.Lock()
	room = NewRoom(id, convo.Name)
	h.Rooms[id] = room
	h.mu.Unlock()
	go room.run()
	return room
}

func (h *Hub) AddClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Clients[client.ID] = client
}

func (h *Hub) DeleteClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.Clients, client.ID)
}

func (h *Hub) GetClient(clientID string) (*Client, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, ok := h.Clients[clientID]

	if !ok {
		return nil, fmt.Errorf("client doesnt exist")
	}
	return client, nil
}

func (h *Hub) JoinRoom(ctx context.Context, client *Client, room *Room) (*Message, error) {
	message := &Message{
		Type:     "join_room",
		SenderType: "system",
		Content:  "a new user has joined the chat",
		Username: client.Username,
		RoomID:   room.ID,
	}

	client.Rooms[room.ID] = room
	room.Register <- client
	clientID64, err := strconv.ParseInt(client.ID, 10, 64)
	if err != nil {
		log.Panic("could not convert string to in64")
		return nil, fmt.Errorf("could not convert string to in64")
	}
	roomID64, err := strconv.ParseInt(room.ID, 10, 64)
	if err != nil {
		log.Panic("could not convert string to in64")
		return nil, fmt.Errorf("could not convert string to in64")
	}
	
	isParticipant, err :=  h.ConvoService.IsParticipant(ctx, roomID64, clientID64)
	if err != nil {
		log.Panic("not a participant of a room")
		return nil, fmt.Errorf("not a participant of a room")
	}
	
	if !isParticipant {
		h.ConvoService.AddParticipant(ctx, roomID64, clientID64)
	}

	room.Broadcast <- message

	return message, nil
}

func (h *Hub) LeaveRoom(client *Client, room *Room) (*Message, error) {
	
	h.mu.Lock()
	message := &Message{
		Type:     "leave_room",
		Content:  "a new user has left the chat",
		SenderType: "system",
		Username: client.Username,
		RoomID:   room.ID,
	}

	room, ok := client.Rooms[room.ID]
	if !ok {
		log.Print("room is not registered in client")
		return nil, fmt.Errorf("room is not registered in client")
	}

	delete(client.Rooms, room.ID)
	h.mu.Unlock()
	room.Unregister <- client

	room.Broadcast <- message

	return message, nil
}

func(h *Hub) SendToAi(ctx context.Context, msg *models.Message) (*Message, error) {
	DBMessages, err := h.MessageService.MessageToAi(ctx, msg)
	if err != nil {
		fmt.Println("couldnt send message to AI")
		return nil, fmt.Errorf("couldnt send message to AI: %w", err)
	}

	fmt.Println(DBMessages)

	wsMessage := Message{
		Type: DBMessages.Type ,
		Content: DBMessages.Content,
		Username: "assistant", 
		SenderType: DBMessages.SenderType,
		RoomID: DBMessages.ConvoID,
		UserID: DBMessages.UserID,
	}
	fmt.Println("OBSERVE HERE")
	fmt.Println(wsMessage)
	fmt.Println("OBSERVE HERE")
	return &wsMessage, nil
}