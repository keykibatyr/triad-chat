package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	chatmodels "github.com/keykibatyr/triad-chat/internal/chat/models"
	coremodels "github.com/keykibatyr/triad-chat/internal/models"
)

type ConvoRepository interface {
	CreateConvo(ctx context.Context, convo *chatmodels.Convo) (*chatmodels.Convo, error)
	GetConvoByID(ctx context.Context, id int64) (*chatmodels.Convo, error)
	GetConvoListByName(ctx context.Context, name string) ([]chatmodels.Convo, error)
	AddParticipant(ctx context.Context, convoID, userID int64) error
	RemoveParticipant(ctx context.Context, convoID, userID int64) error
	IsParticipant(ctx context.Context, convoID, userID int64) (bool, error)
	GetConversationParticipants(ctx context.Context, convoID int64) ([]coremodels.User, error)
	GetUserConversationIDs(ctx context.Context, userID int64) ([]int64, error)
}

func NewConvoRepo(db *sql.DB) ConvoRepository {
	return &PostgresConvoRepository{
		DB: db,
	}
}

type PostgresConvoRepository struct {
	DB *sql.DB
}

func (r *PostgresConvoRepository) CreateConvo(ctx context.Context, convo *chatmodels.Convo) (*chatmodels.Convo, error) {
	query := `INSERT INTO conversations (conversation_type, name) VALUES ($1, $2) RETURNING id`

	err := r.DB.QueryRowContext(ctx, query, convo.Type, convo.Name).Scan(&convo.ID)
	if err != nil {
		return nil, fmt.Errorf("fail at inserting conversation: %w", err)
	}

	return convo, nil
}
func (r *PostgresConvoRepository) GetConvoByID(ctx context.Context, id int64) (*chatmodels.Convo, error) {
	var convo chatmodels.Convo

	convo.ID = id

	query := `SELECT conversation_type, name FROM conversations WHERE id = $1`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(&convo.Type, &convo.Name)
	if err != nil {
		return nil, fmt.Errorf("fail at selecting Conversation: %w", err)
	}

	return &convo, nil
}

func (r *PostgresConvoRepository) AddParticipant(ctx context.Context, convoID, userID int64) error {
	query := `INSERT INTO conversation_participants (conversation_id, user_id) VALUES ($1, $2)`

	err := r.DB.QueryRowContext(ctx, query, convoID, userID)
	if err != nil {
		return fmt.Errorf("fail at setting relationship user x conversation: %v", err)
	}

	return nil
}

func (r *PostgresConvoRepository) RemoveParticipant(ctx context.Context, convoID, userID int64) error {
	query := `DELETE FROM conversation_participants WHERE conversation_id = $1 AND user_id = $2`

	err := r.DB.QueryRowContext(ctx, query, convoID, userID)
	if err != nil {
		return fmt.Errorf("fail at deleting relationship user x conversation: %v", err)
	}

	return nil
}

func (r *PostgresConvoRepository) IsParticipant(ctx context.Context, convoID, userID int64) (bool, error) {
	query := `SELECT 1 FROM conversation_participants WHERE conversation_id = $1 AND user_id = $2`

	var exists int
	err := r.DB.QueryRowContext(ctx, query, convoID, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check participant: %w", err)
	}

	return true, nil
}

func (r *PostgresConvoRepository) GetConversationParticipants(ctx context.Context, convoID int64) ([]coremodels.User, error) {
	var users []coremodels.User

	query := `SELECT id, email, password_hash, role, username FROM users u 
	INNER JOIN conversation_participants c 
	ON u.id = c.user_id WHERE c.conversation_id = $1)`

	rows, err := r.DB.QueryContext(ctx, query, convoID)
	if err != nil {
		return nil, fmt.Errorf("could not get users of the conversation: %w", err)
	}

	for rows.Next(){
		var user coremodels.User

		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Username)
		if err != nil {
			return nil, fmt.Errorf("could not build user of the conversation: %w", err)
		}

		users = append(users, user)
		
	}

	return users, nil
}

func (r *PostgresConvoRepository) GetUserConversationIDs(ctx context.Context, userID int64) ([]int64, error) {
	var convoIDs []int64

	query := `SELECT id FROM conversation_participants WHERE user_id = $1`

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("could not get the conversations of the users: %w", err)
	}

	for rows.Next(){
		var convoID int64

		err := rows.Scan(&convoID)
		if err != nil {
			return nil, fmt.Errorf("could not get the conversations of the user: %w", err)
		}

		convoIDs = append(convoIDs, convoID)
	}

	return  convoIDs, nil
}


func (r *PostgresConvoRepository) GetConvoListByName(ctx context.Context, name string) ([]chatmodels.Convo, error) {
	var convoList []chatmodels.Convo

	query := `SELECT id, name, conversation_type FROM conversations WHERE name = $1`

	rows, err := r.DB.QueryContext(ctx, query, name)
	if err != nil {
		return nil, fmt.Errorf("could not get the conversations ids by name: %w", err)
	}

	for rows.Next(){
		var convo chatmodels.Convo

		err := rows.Scan(&convo.ID, &convo.Name, &convo.Type)
		if err != nil {
			return nil, fmt.Errorf("could not get the conversations of the user: %w", err)
		}

		convoList = append(convoList, convo)
	}

	return  convoList, nil
}