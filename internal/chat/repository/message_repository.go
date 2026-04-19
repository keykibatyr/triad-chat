package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/keykibatyr/triad-chat/internal/chat/models"
)

type MessageRepository interface {
	CreateMessage(ctx context.Context, message *models.Message) (*models.Message, error)
	GetByID(ctx context.Context, ID int64) (*models.Message, error)
	GetByConvoID(ctx context.Context, convoID int64) ([]*models.Message, error)
	GetByUserID(ctx context.Context, userID int64) ([]*models.Message, error)
}

type PostgresMessageRepository struct {
	DB *sql.DB
}

func NewMessageRepo(db *sql.DB) MessageRepository {
	return &PostgresMessageRepository{
		DB: db,
	}
}

func (r *PostgresMessageRepository) CreateMessage(ctx context.Context, message *models.Message) (*models.Message, error) {
	query := `INSERT INTO messages (type, conversation_id, sender_id, content) VALUES ($1, $2, $3, $4) RETURNING id`

	err := r.DB.QueryRowContext(ctx, query,
		message.Type, 
		message.ConvoID, 
		message.UserID, 
		message.Content).Scan(&message.ID)

	if err != nil {
		return nil, fmt.Errorf("could not insert a message: %w", err)
	}

	return message, nil
}

func (r *PostgresMessageRepository) GetByID(ctx context.Context, id int64) (*models.Message, error) {
	var message models.Message
	query := `SELECT (conversation_id, type, sender_id, content FROM messages WHERE id = $1`

		err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&message.ID,
		&message.ConvoID,
		&message.Type,
		&message.UserID,
		&message.Content,
	)

	if err != nil {
		return nil, fmt.Errorf("could not select a message: %w", err)
	}

	return &message, nil
}

func (r *PostgresMessageRepository) GetByConvoID(ctx context.Context, convoID int64) ([]*models.Message, error) {
	var messages []*models.Message

	query := `SELECT (id, type, sender_id, content FROM messages WHERE conversation_id = $1`

	rows, err := r.DB.QueryContext(ctx, query, convoID)
	if err != nil {
		return nil, fmt.Errorf("failed scanning rows: %w", err)
	}

	for rows.Next() {
		var message models.Message
		err := rows.Scan(
			&message.ID, 
			&message.Type, 
			&message.UserID, 
			&message.Content,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning a row: %w", err)
		}
		message.ConvoID = strconv.FormatInt(convoID, 10)

		messages = append(messages, &message)
	}

	defer rows.Close()

	return messages, nil
}

func (r *PostgresMessageRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.Message, error) {
	var messages []*models.Message

	query := `SELECT (id, conversation_id, type, content FROM messages WHERE sender_id = $1`

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed scanning rows: %w", err)
	}

	for rows.Next() {
		var message models.Message
		err := rows.Scan(
			&message.ID, 
			&message.ConvoID, 
			&message.Type, 
			&message.Content,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning a row: %w", err)
		}
		message.UserID = strconv.FormatInt(userID, 10)
		messages = append(messages, &message)
	}

	defer rows.Close()

	return messages, nil
}
