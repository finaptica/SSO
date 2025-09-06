package services

import (
	"context"
	"time"

	"github.com/finaptica/sso/internal/domain/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email string, passHash []byte) (uid uuid.UUID, err error)
	GetUserByIDTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
}

type AppRepository interface {
	GetApp(ctx context.Context, appId int) (models.App, error)
	GetAppByIDTx(ctx context.Context, tx *sqlx.Tx, id int) (models.App, error)
}

type RefreshTokenRepository interface {
	SaveNewRefreshToken(ctx context.Context, userId uuid.UUID, appId int, value string, expiresAt time.Time) error
	SaveNewRefreshTokenTx(ctx context.Context, tx *sqlx.Tx, userId uuid.UUID, appId int, value string, expiresAt time.Time) error
	RevokeTx(ctx context.Context, tx *sqlx.Tx, tokenID uuid.UUID) error
	GetByValueTx(ctx context.Context, tx *sqlx.Tx, tokenValue string) (models.RefreshToken, error)
	GetByValue(ctx context.Context, tokenValue string) (*models.RefreshToken, error)
	DB() *sqlx.DB
}
