package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/keykibatyr/triad-chat/internal/chat/models"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	Rooms    map[string]*Room
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Message struct {
	Type     string `json:"type"`
	Content  string `json:"content"`
	Username string `json:"username"`
	RoomID   string `json:"room_id"`
	UserID   string `json:"user_id"`
}

func (c *Client) readPump() {
	defer func() {
		for _, room := range c.Rooms {
			room.Unregister <- c
		}
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("error parsing message: %v", err)
			continue
		}

		msg.Username = c.Username
		msg.UserID = c.ID

		messageDB := &models.Message{
			Type:    msg.Type,
			Content: msg.Content,
			ConvoID: msg.RoomID,
			UserID:  msg.UserID,
		}

		switch msg.Type {
		case "join_room": //fix
			room := c.Hub.GetRoom(msg.RoomID)
			if room == nil {
				continue
			}
			c.Hub.JoinRoom(c, room)
			c.Hub.Worker.Enqueue(messageDB)
			log.Println("message sent to db 1")

		case "leave_room": //fix
			room, ok := c.Rooms[msg.RoomID]
			if !ok {
				continue
			}
			c.Hub.LeaveRoom(c, room)
			c.Hub.Worker.Enqueue(messageDB)
			log.Println("message sent to db 2")

		case "message":
			room, ok := c.Rooms[msg.RoomID]
			if !ok {
				continue
			}

			room.Broadcast <- &msg
			c.Hub.Worker.Enqueue(messageDB)
			log.Println("message sent to db 3")
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.Hub.DeleteClient(c)
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			err = w.Close()
			if err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
