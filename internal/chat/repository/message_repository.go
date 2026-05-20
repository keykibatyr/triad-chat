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
	GetByConvoID(ctx context.Context, convoID int64) ([]models.Message, error)
	GetByUserID(ctx context.Context, userID int64) ([]models.Message, error)
	GetMeesages(ctx context.Context, convoID int64, messageLimit int) ([]models.Message, int, error)
	GetMessagesAndCursor(ctx context.Context, convoID int64, messageLimit, cursor int) ([]models.Message, int, error)
	GetXMeesages(ctx context.Context, convoID int64, messageLimit int) ([]models.Message, error)
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
	query := `INSERT INTO messages (type, conversation_id, sender_id, content, sender_type) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`

	err := r.DB.QueryRowContext(ctx, query,
		message.Type,
		message.ConvoID,
		message.UserID,
		message.Content,
		message.SenderType).Scan(&message.ID, &message.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("could not insert a message: %w", err)
	}

	return message, nil
}

func (r *PostgresMessageRepository) GetByID(ctx context.Context, id int64) (*models.Message, error) {
	var message models.Message
	query := `SELECT conversation_id, type, sender_type, sender_id, content, created_at FROM messages WHERE id = $1`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&message.ID,
		&message.ConvoID,
		&message.Type,
		&message.SenderType,
		&message.UserID,
		&message.Content,
		&message.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("could not select a message: %w", err)
	}

	return &message, nil
}

func (r *PostgresMessageRepository) GetByConvoID(ctx context.Context, convoID int64) ([]models.Message, error) {
	var messages []models.Message

	query := `SELECT id, type, sender_type, sender_id, content, created_at FROM messages WHERE conversation_id = $1`

	rows, err := r.DB.QueryContext(ctx, query, convoID)
	if err != nil {
		return nil, fmt.Errorf("failed scanning rows: %w", err)
	}

	for rows.Next() {
		var message models.Message
		err := rows.Scan(
			&message.ID,
			&message.Type,
			&message.SenderType,
			&message.UserID,
			&message.Content,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning a row: %w", err)
		}
		message.ConvoID = strconv.FormatInt(convoID, 10)

		messages = append(messages, message)
	}

	defer rows.Close()

	return messages, nil
}

func (r *PostgresMessageRepository) GetByUserID(ctx context.Context, userID int64) ([]models.Message, error) {
	var messages []models.Message

	query := `SELECT id, conversation_id, type, sender_type, content, created_at FROM messages WHERE sender_id = $1`

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
			&message.SenderType,
			&message.Content,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning a row: %w", err)
		}
		
		messages = append(messages, message)
	}

	defer rows.Close()

	return messages, nil
}

func (r *PostgresMessageRepository) GetMeesages(ctx context.Context, convoID int64, messageLimit int) ([]models.Message, int, error) {
	var messages []models.Message
	var cursor int

	query := `SELECT id, conversation_id, sender_id, type, sender_type, content, created_at 
	FROM messages WHERE conversation_id = $1 
	ORDER BY created_at DESC, id DESC LIMIT $2`

	rows, err := r.DB.QueryContext(ctx, query, convoID, messageLimit)
	if err != nil {
		return nil, 0, fmt.Errorf("fail retrieving X messages from DB: %w", err)
	}

	for rows.Next() {
		var message models.Message

		err := rows.Scan(
			&message.ID,
			&message.ConvoID,
			&message.UserID,
			&message.Type,
			&message.SenderType,
			&message.Content,
			&message.CreatedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("fail assigning single message: %w", err)
		}

		messages = append(messages, message)

	}

	return messages, cursor, nil
}


func (r *PostgresMessageRepository) GetMessagesAndCursor(ctx context.Context, convoID int64, messageLimit, cursor int) ([]models.Message, int, error){
	var messages []models.Message

	query := `SELECT id, conversation_id, sender_id, type, sender_type, content, created_at 
	FROM messages WHERE conversation_id = $1 
	AND id < $2
	ORDER BY created_at DESC, id DESC LIMIT $3`

	rows, err := r.DB.QueryContext(ctx, query, convoID, cursor, messageLimit)
	if err != nil {
		return nil, 0, fmt.Errorf("fail retrieving X messages from DB: %w", err)
	}

	for rows.Next() {
		var message models.Message

		err := rows.Scan(
			&message.ID,
			&message.ConvoID,
			&message.UserID,
			&message.Type,
			&message.SenderType,
			&message.Content,
			&message.CreatedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("fail assigning single message: %w", err)
		}

		messages = append(messages, message)

	}

	return messages, cursor, nil
}

func (r *PostgresMessageRepository) GetXMeesages(ctx context.Context, convoID int64, messageLimit int) ([]models.Message, error) {
	var messages []models.Message

	query := `SELECT id, conversation_id, sender_id, type, sender_type, content, created_at 
	FROM messages WHERE conversation_id = $1 
	ORDER BY created_at DESC, id DESC LIMIT $2`

	rows, err := r.DB.QueryContext(ctx, query, convoID, messageLimit)
	if err != nil {
		return nil, fmt.Errorf("fail retrieving X messages from DB: %w", err)
	}

	for rows.Next() {
		var message models.Message

		err := rows.Scan(
			&message.ID,
			&message.ConvoID,
			&message.UserID,
			&message.Type,
			&message.SenderType,
			&message.Content,
			&message.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("fail assigning single message: %w", err)
		}

		messages = append(messages, message)

	}

	return messages, nil
}
