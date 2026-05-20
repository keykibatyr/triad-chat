package models

type AiMessage struct {
	Role string `json:"role"`
	Content string `json:"content"`
}

type AiRequest struct {
	Messages []AiMessage `json:"ai_messages"`
	SystemText string
}