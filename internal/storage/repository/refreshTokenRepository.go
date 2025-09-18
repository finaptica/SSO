package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/finaptica/sso/internal/domain/models"
	"github.com/finaptica/sso/internal/lib/errs"
	"github.com/finaptica/sso/internal/lib/logger/sl"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RefreshTokenRepository struct {
	db  *pgx.Conn
	log *slog.Logger
}

func NewRefreshTokenRepository(log *slog.Logger, db *pgx.Conn) *RefreshTokenRepository {
	return &RefreshTokenRepository{log: log, db: db}
}

func (r *RefreshTokenRepository) SaveNewRefreshToken(ctx context.Context, userId uuid.UUID, appId int, value string, expiresAt time.Time) (uuid.UUID, error) {
	const op = "refreshTokenRepository.SaveNewRefreshToken"
	log := r.log.With(slog.String("op", op), slog.String("refreshTokenValue", value))
	var id uuid.UUID
	err := r.db.QueryRow(ctx, "INSERT INTO refresh_tokens (user_id, app_id, value, created_at, expires_at) VALUES ($1, $2, $3, $4, $5) RETURNING id", userId, appId, value, time.Now().UTC(), expiresAt).Scan(&id)
	if err != nil {
		log.Error("failed to create refresh token", sl.Err(err))
		return uuid.UUID{}, errs.WithKind(op, errs.Internal, err)
	}

	return id, nil
}

func (r *RefreshTokenRepository) SaveNewRefreshTokenTx(ctx context.Context, tx pgx.Tx, userId uuid.UUID, appId int, value string, expiresAt time.Time) (uuid.UUID, error) {
	const op = "refreshTokenRepository.SaveNewRefreshTokenTx"
	log := r.log.With(slog.String("op", op), slog.String("refreshTokenValue", value))

	var id uuid.UUID
	err := tx.QueryRow(
		ctx,
		`INSERT INTO refresh_tokens (user_id, app_id, value, created_at, expires_at) 
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userId, appId, value, time.Now().UTC(), expiresAt,
	).Scan(&id)

	if err != nil {
		log.Error("failed to create refresh token", sl.Err(err))
		return uuid.UUID{}, errs.WithKind(op, errs.Internal, err)
	}

	return id, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, tokenId uuid.UUID) error {
	const op = "refreshTokenRepository.Revoke"
	log := r.log.With(slog.String("op", op))
	_, err := r.db.Exec(ctx, "UPDATE refresh_tokens SET is_revoked = true WHERE id = $1", tokenId)
	if err != nil {
		log.Error("failed to revoke token", sl.Err(err))
		return errs.WithKind(op, errs.Internal, err)
	}

	return nil
}

func (r *RefreshTokenRepository) GetByValue(ctx context.Context, tokenValue string) (*models.RefreshToken, error) {
	const op = "refreshTokenRepository.GetByValue"
	log := r.log.With(slog.String("op", op))
	var token models.RefreshToken
	err := r.db.QueryRow(ctx, "SELECT id, user_id, app_id, value, is_revoked, created_at, expires_at FROM refresh_tokens WHERE value = $1", tokenValue).Scan(&token.ID, &token.UserID, &token.AppID, &token.Value, &token.IsRevoked, &token.CreatedAt, &token.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return &models.RefreshToken{}, errs.WithKind(op, errs.NotFound, err)
		}
		log.Error("failed to get token by value", sl.Err(err))
		return &models.RefreshToken{}, errs.WithKind(op, errs.Internal, err)
	}

	return &token, nil
}

func (r *RefreshTokenRepository) RevokeTx(ctx context.Context, tx pgx.Tx, tokenID uuid.UUID) error {
	_, err := tx.Exec(ctx, "UPDATE refresh_tokens SET is_revoked = true WHERE id = $1", tokenID)
	return err
}

func (r *RefreshTokenRepository) GetByValueTx(ctx context.Context, tx pgx.Tx, tokenValue string) (models.RefreshToken, error) {
	var token models.RefreshToken
	err := tx.QueryRow(ctx, "SELECT id, user_id, app_id, value, is_revoked, created_at, expires_at  FROM refresh_tokens WHERE value = $1", tokenValue).Scan(&token.ID, &token.UserID, &token.AppID, &token.Value, &token.IsRevoked, &token.CreatedAt, &token.ExpiresAt)
	return token, err
}
