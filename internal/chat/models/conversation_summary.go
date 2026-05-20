package models

type ConvoSummary struct {
	ID int64 `json:"id"`
	ConvoID string `json:"convo_id"`
	SummaryText string `json:"summary_text"`
	LastMessageID string `json:"last_message_id"`
}

