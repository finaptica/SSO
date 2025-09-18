package repository

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/finaptica/sso/internal/domain/models"
	"github.com/finaptica/sso/internal/lib/errs"
	"github.com/finaptica/sso/internal/lib/logger/sl"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db  *pgx.Conn
	log *slog.Logger
}

func NewUserRepository(log *slog.Logger, db *pgx.Conn) *UserRepository {
	return &UserRepository{log: log, db: db}
}

func (u *UserRepository) CreateUser(ctx context.Context, email string, passHash []byte) (uuid.UUID, error) {
	const op = "userRepository.CreateUser"
	log := u.log.With(slog.String("op", op), slog.String("email", email))
	var id uuid.UUID
	err := u.db.QueryRow(ctx, "INSERT INTO users (email, pass_hash) VALUES ($1, $2) RETURNING id", email, passHash).Scan(&id)
	if err != nil {
		log.Error("failed to create user", sl.Err(err))
		return uuid.UUID{}, errs.WithKind(op, errs.Internal, err)
	}

	return id, nil
}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	const op = "userRepository.GetUserByEmail"
	log := u.log.With(slog.String("op", op), slog.String("email", email))
	var user models.User
	err := u.db.QueryRow(ctx,
		`SELECT id, email, pass_hash FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Email, &user.PassHash)

	if err != nil {
		if err == pgx.ErrNoRows {
			log.Info("user not found")
			return models.User{}, errs.WithKind(op, errs.NotFound, err)
		}

		log.Error("failed to get user by email", sl.Err(err))
		return models.User{}, errs.WithKind(op, errs.Internal, err)
	}
	return user, nil
}

func (u *UserRepository) GetUserByIDTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (models.User, error) {
	const op = "userRepository.GetUserByIDTx"
	log := u.log.With(slog.String("op", op), slog.String("userID", id.String()))

	var user models.User
	err := tx.QueryRow(ctx, `SELECT id, email, pass_hash FROM users WHERE id = $1`, id).Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("user not found")
			return models.User{}, errs.WithKind(op, errs.NotFound, err)
		}
		log.Error("failed to get user by ID", sl.Err(err))
		return models.User{}, errs.WithKind(op, errs.Internal, err)
	}

	return user, nil
}

func (u *UserRepository) IsUserExistByEmail(ctx context.Context, email string) (bool, error) {
	const op = "userRepository.IsUserExistByEmail"
	log := u.log.With(slog.String("op", op), slog.String("email", email))
	var isExist bool
	err := u.db.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)", email).Scan(&isExist)
	if err != nil {
		log.Error("failed to check exist user or not", sl.Err(err))
		return false, errs.WithKind(op, errs.Internal, err)
	}

	return isExist, nil
}
