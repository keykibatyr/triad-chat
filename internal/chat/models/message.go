package models

import "time"

type Message struct {
	ID         int64     `json:"id"`
	Type       string    `json:"type"`
	SenderType string    `json:"sender_type"`
	Content    string    `json:"content"`
	ConvoID    string    `json:"convo_id"`
	UserID     string    `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
}
