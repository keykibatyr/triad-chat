package ws

import (
	"fmt"
	"log"
	"sync"

	"github.com/keykibatyr/triad-chat/internal/chat/worker"
)

type Hub struct {
	Rooms   map[string]*Room
	Clients map[string]*Client
	mu      sync.RWMutex

	Worker *worker.PersistWorker
}

func NewHub(Worker *worker.PersistWorker) *Hub {
	return &Hub{
		Rooms:   make(map[string]*Room),
		Clients: make(map[string]*Client),

		Worker: Worker,
	}
}

func (h *Hub) GetOrCreateRoom(id, name string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	room, ok := h.Rooms[id]
	if ok {
		return room
	}

	room = NewRoom(id, name)
	h.Rooms[id] = room
	go room.run()
	return room
}

func (h *Hub) GetRoom(id string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	room, ok := h.Rooms[id]
	if ok {
		return room
	}

	room = NewRoom(id, "test")
	h.Rooms[id] = room
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

func (h *Hub) JoinRoom(client *Client, room *Room) {

	message := &Message{
		Type:     "join_room",
		Content:  "a new user has joined the chat",
		Username: client.Username,
		RoomID:   room.ID,
	}

	client.Rooms[room.ID] = room
	room.Register <- client

	room.Broadcast <- message
}

func (h *Hub) LeaveRoom(client *Client, room *Room) {

	message := &Message{
		Type:     "leave_room",
		Content:  "a new user has left the chat",
		Username: client.Username,
		RoomID:   room.ID,
	}

	room, ok := client.Rooms[room.ID]
	if !ok {
		log.Print("room is not registered in client")
		return
	}

	delete(client.Rooms, room.ID)
	room.Unregister <- client

	room.Broadcast <- message
}
