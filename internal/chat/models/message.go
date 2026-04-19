package models

type Message struct {
	ID int64 `json:"id"`
	Type     string	`json:"type"`
	Content  string `json:"content"`
	ConvoID   string `json:"convo_id"`
	UserID string 	`json:"user_id"`
}
