package services

import (
	"context"
	"time"

	"github.com/finaptica/sso/internal/domain/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email string, passHash []byte) (uid uuid.UUID, err error)
	GetUserByIDTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
}

type AppRepository interface {
	GetAppById(ctx context.Context, appId int) (models.App, error)
	GetAppByIDTx(ctx context.Context, tx pgx.Tx, id int) (models.App, error)
}

type RefreshTokenRepository interface {
	SaveNewRefreshToken(ctx context.Context, userId uuid.UUID, appId int, value string, expiresAt time.Time) (uuid.UUID, error)
	SaveNewRefreshTokenTx(ctx context.Context, tx pgx.Tx, userId uuid.UUID, appId int, value string, expiresAt time.Time) (uuid.UUID, error)
	RevokeTx(ctx context.Context, tx pgx.Tx, tokenID uuid.UUID) error
	GetByValueTx(ctx context.Context, tx pgx.Tx, tokenValue string) (models.RefreshToken, error)
	GetByValue(ctx context.Context, tokenValue string) (*models.RefreshToken, error)
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(pgx.Tx) error) error
}

type RepositoriesContainer struct {
	UserRepo UserRepository
	AppRepo  AppRepository
	RtsRepo  RefreshTokenRepository
	Uow      UnitOfWork
}
