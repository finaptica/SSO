package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/finaptica/sso/internal/domain/models"
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

func (u *UserRepository) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "UserRepository.SaveUser"
	u.log = u.log.With(slog.String("op", op))
	var id int64
	err := u.db.QueryRowxContext(ctx, "INSERT INTO users (email, pass_hash) VALUES ($1, $2) RETURNING id", email, passHash).Scan(&id)
	if err != nil {
		u.log.Info("failed to create useer")
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *UserRepository) User(ctx context.Context, email string) (models.User, error) {
	var u models.User
	err := r.db.GetContext(ctx, &u,
		`SELECT id, email, pass_hash FROM users WHERE email = $1`, email,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, storage.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

func (r *UserRepository) IsAdmin(ctx context.Context, userId int64) (bool, error) {
	var isAdmin bool
	err := r.db.GetContext(ctx, &isAdmin,
		`SELECT is_admin FROM users WHERE id = $1`, userId,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, storage.ErrUserNotFound
		}
		return false, fmt.Errorf("check admin: %w", err)
	}
	return isAdmin, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && err.Error() != "" &&
		(len(err.Error()) >= 21 && err.Error()[0:21] == "pq: duplicate key")
}
