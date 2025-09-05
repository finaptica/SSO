package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/finaptica/sso/internal/lib/errs"
	"github.com/finaptica/sso/internal/lib/logger/sl"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RefreshTokenRepository struct {
	db  *sqlx.DB
	log *slog.Logger
}

func NewRefreshTokenRepository(log *slog.Logger, db *sqlx.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{log: log, db: db}
}

func (r *RefreshTokenRepository) SaveNewRefreshToken(ctx context.Context, userId uuid.UUID, appId int, value string, expiresAt time.Time) error {
	const op = "refreshTokenRepository.SaveNewRefreshToken"
	log := r.log.With(slog.String("op", op), slog.String("refreshTokenValue", value))
	var id uuid.UUID
	err := r.db.QueryRowxContext(ctx, "INSERT INTO refresh_tokens (user_id, app_id, value, created_at, expires_at) VALUES ($1, $2, $3, $4, $5) RETURNING id", userId, appId, value, time.Now().UTC(), expiresAt).Scan(&id)
	if err != nil {
		log.Error("failed to create refresh token", sl.Err(err))
		return errs.WithKind(op, errs.Internal, err)
	}

	return nil
}
