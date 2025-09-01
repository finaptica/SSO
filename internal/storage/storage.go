package storage

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrAppNotFound  = errors.New("app not found")
)

func New(logger *slog.Logger, postgresConnectionString string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", postgresConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect postgres: %w", err)
	}

	return db, nil
}
