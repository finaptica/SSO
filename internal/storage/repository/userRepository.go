package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/finaptica/sso/internal/domain/models"
	"github.com/finaptica/sso/internal/lib/logger/sl"
	"github.com/finaptica/sso/internal/storage"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db  *sqlx.DB
	log *slog.Logger
}

func NewUserRepository(log *slog.Logger, db *sqlx.DB) *UserRepository {
	return &UserRepository{log: log, db: db}
}

func (u *UserRepository) CreateUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "userRepository.CreateUser"
	log := u.log.With(slog.String("op", op), slog.String("email", email))
	var id int64
	err := u.db.QueryRowxContext(ctx, "INSERT INTO users (email, pass_hash) VALUES ($1, $2) RETURNING id", email, passHash).Scan(&id)
	if err != nil {
		log.Error("failed to create useer", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	const op = "userRepository.GetUserByEmail"
	log := u.log.With(slog.String("op", op), slog.String("email", email))
	var user models.User
	err := u.db.GetContext(ctx, &user,
		`SELECT id, email, pass_hash FROM users WHERE email = $1`, email,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("user not found")
			return models.User{}, storage.ErrUserNotFound
		}

		log.Error("failed to get user by email", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func (u *UserRepository) IsUserExistByEmail(ctx context.Context, email string) (bool, error) {
	const op = "userRepository.IsUserExistByEmail"
	log := u.log.With(slog.String("op", op), slog.String("email", email))
	var isExist bool
	err := u.db.GetContext(ctx, &isExist,
		"SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)", email)
	if err != nil {
		log.Error("failed to check exist user or not", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isExist, nil
}
