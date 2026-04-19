package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/keykibatyr/triad-chat/internal/models"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	DeleteById(ctx context.Context, jti uuid.UUID) error
	DeleteByUserId(ctx context.Context, userID int64) error
}

func NewRefreshRepo(db *sql.DB) RefreshTokenRepository {
	return &PostgresRefreshToken{
		DB: db,
	}
}

type PostgresRefreshToken struct {
	DB *sql.DB
}

func (r *PostgresRefreshToken) Create(ctx context.Context, token *models.RefreshToken) error {

	query := `INSERT INTO refresh_tokens (user_id, token_hash, expires_at, jti) 
	VALUES ($1, $2, $3, $4) RETURNING id`

	err := r.DB.QueryRowContext(ctx, query, token.UserID, token.TokenHash, token.ExpiresAt, token.JTI).
		Scan(&token.ID)

	if err != nil {
		return fmt.Errorf("could not insert a Refresh_token")
	}

	return nil
}

func (r *PostgresRefreshToken) GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error){
	var token models.RefreshToken

	query := `SELECT id, user_id, token_hash, expires_at, jti FROM refresh_tokens 
	WHERE token_hash = $1`

	err := r.DB.QueryRowContext(ctx, query, tokenHash).
	Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.JTI)

	if err != nil {
		return nil, fmt.Errorf("could not get the Token by Hash")
	}

	return &token, nil
}

func (r *PostgresRefreshToken) DeleteById(ctx context.Context, jti uuid.UUID) error {

	query := `DELETE FROM refresh_tokens WHERE jti = $1`

	_, err := r.DB.ExecContext(ctx, query, jti)
	if err != nil {
		return fmt.Errorf("could not delete RefreshToken by ID")
	}

	return nil
}

func (r *PostgresRefreshToken) DeleteByUserId(ctx context.Context, userID int64) error {

	query := `DELERE FROM refresh_token WHERE user_id = $1`

	_, err := r.DB.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("could not delete RefreshToken by userID")
	}

	return nil
}