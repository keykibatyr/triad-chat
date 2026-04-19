package repository

//REFACTOR ERRORS

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/keykibatyr/triad-chat/internal/auth"
	"github.com/keykibatyr/triad-chat/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetById(ctx context.Context, id int64) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

type AuthService interface {
	Register(ctx context.Context, email, password, username string) (*models.User, *auth.TokenPair, error)
	Login(ctx context.Context, email, password string) (*models.User, *auth.TokenPair, error)
	RefreshToken(ctx context.Context, refreshTokenString string) (*auth.TokenPair, error)
	Logout(ctx context.Context, refreshTOkenString string) (error)
}

type UserService interface {
	GetByID(ctx context.Context, id int64) (*models.User, error)
}

type PostgresUserRepository struct {
	DB *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepository {
	return &PostgresUserRepository{
		DB: db,
	}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {

	query := `INSERT INTO users (email, password_hash, role, username) 
	VALUES ($1, $2, $3, $4) RETURNING id`

	err := r.DB.QueryRowContext(ctx, query, user.Email, user.PasswordHash, user.Role, user.Username).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("could not insert a user")
	}

	return nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	query := `SELECT id, email, password_hash, role, username
	FROM users WHERE email = $1`

	err := r.DB.QueryRowContext(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Username)

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("could not get user by Email")
	}

	return &user, nil
}

func (r *PostgresUserRepository) GetById(ctx context.Context, id int64) (*models.User, error) {
	var user models.User

	query := `SELECT id, email, password_hash, role, username
	FROM users WHERE id = $1`

	err := r.DB.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Username)

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("could not get user by ID")
	}

	return &user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *models.User) error {

	query := `UPDATE users 
	SET email = $1, password_hash = $2, role = $3, username = $4
	WHERE id = $5`

	_, err := r.DB.ExecContext(ctx, query,
		user.Email, user.PasswordHash, user.Role, user.Username, user.ID)

	if err != nil {
		return fmt.Errorf("could not update the User")
	}

	return nil
}
