package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/keykibatyr/triad-chat/internal/chat/models"
)

type ConvoSummaryRepository interface {
	CreateConvoSummary(ctx context.Context, convoSummary *models.ConvoSummary) (*models.ConvoSummary, error)
	GetConvoSummaryByConvoID(ctx context.Context, convoID int64) (*models.ConvoSummary, error)
	UpdateConvoSummary(ctx context.Context, convoID, lastMessageID int64, summaryText string) error
}

func NewConvoSummaryRepo(db *sql.DB) ConvoSummaryRepository {
	return &PostgresConvoSummaryRepository{
		DB: db,
	}
}

type PostgresConvoSummaryRepository struct {
	DB *sql.DB
}

func (r *PostgresConvoSummaryRepository) CreateConvoSummary(ctx context.Context, convoSummary *models.ConvoSummary) (*models.ConvoSummary, error) {
	query := `INSERT INTO conversation_summaries (conversation_id, summary_text, last_message_id) VALUES ($1, $2, $3) RETURNING id`

	err := r.DB.QueryRowContext(ctx, query, convoSummary.ConvoID, convoSummary.SummaryText, convoSummary.LastMessageID).Scan(&convoSummary.ID)
	if err != nil {
		return nil, fmt.Errorf("fail at inserting conversation summary: %w", err)
	}

	return convoSummary, nil
}

func (r *PostgresConvoSummaryRepository) GetConvoSummaryByConvoID(ctx context.Context, convoID int64) (*models.ConvoSummary, error) {
	var convoSummary models.ConvoSummary

	query := `SELECT id, conversation_id, summary_text, last_message_id FROM conversation_summaries WHERE conversation_id = $1`

	err := r.DB.QueryRowContext(ctx, query, convoID).Scan(
		&convoSummary.ID,
		&convoSummary.ConvoID,
		&convoSummary.SummaryText,
		&convoSummary.LastMessageID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("could not get summary by convoID: %w", err)
	}

	return &convoSummary, nil
}

func (r *PostgresConvoSummaryRepository) UpdateConvoSummary(ctx context.Context, convoID, lastMessageID int64, summaryText string) error {
	var convoSummary models.ConvoSummary

	query := `UPDATE conversation_summaries SET summary_text = $1, last_message_id = $2 WHERE conversation_id = $3`

	err := r.DB.QueryRowContext(ctx, query, convoID).Scan(
		&convoSummary.ID,
		&convoSummary.ConvoID,
		&convoSummary.SummaryText,
		&convoSummary.LastMessageID,
	)

	if err != nil {
		return fmt.Errorf("could not get the summary by convoID")
	}

	return nil
}
