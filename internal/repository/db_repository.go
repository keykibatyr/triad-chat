package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/keykibatyr/triad-chat/internal/models"
	"github.com/pressly/goose/v3"
)

func Open(cfg models.PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.String())
	if err != nil {
		return nil, fmt.Errorf("opening: %w", err)
	}

	return db, nil
}

func Migrate(db *sql.DB, dir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, dir); err != nil {
		fmt.Println("bug3")
		panic(err)
	}

	return nil
}
