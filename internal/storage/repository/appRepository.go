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

type AppRepository struct {
	db  *sqlx.DB
	log *slog.Logger
}

func NewAppRepository(log *slog.Logger, db *sqlx.DB) *AppRepository {
	return &AppRepository{log: log, db: db}
}

func (r *AppRepository) GetApp(ctx context.Context, appId int) (models.App, error) {
	var app models.App
	err := r.db.GetContext(ctx, &app,
		`SELECT id, name FROM apps WHERE id = $1`, appId,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.App{}, storage.ErrUserNotFound
		}
		return models.App{}, fmt.Errorf("get user: %w", err)
	}
	return app, nil
}
