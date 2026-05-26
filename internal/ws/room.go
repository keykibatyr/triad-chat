package ws

import (
	"encoding/json"
	"log"
	"sync"
)

type RoomReq struct {
	client *Client
	Room   string
}

type Room struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Clients    map[*Client]bool `json:"clients"`
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
	RoomSummary string
}

func NewRoom(id, name string) *Room {
	return &Room{
		ID:         id,
		Name:       name,
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (r *Room) run() {
	for {
		select {
		case client := <-r.Register:
			r.Clients[client] = true
			log.Printf("Client connected: %s", client.Username)
		case client := <-r.Unregister:
			delete(r.Clients, client)
			log.Println(*client, "CLIENT DISSCONECTED: %s", client.Username)
		case message := <-r.Broadcast:
			r.broadcastToRoom(message)
		}
	}
}

func (r *Room) broadcastToRoom(msg *Message) {

	message, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error marshaling message: %v", err)
		return
	}

	for client := range r.Clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(r.Clients, client)
		}
	}
}
