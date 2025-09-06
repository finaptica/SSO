package repository

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/finaptica/sso/internal/domain/models"
	"github.com/finaptica/sso/internal/lib/errs"
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
	const op = "appRepository.GetApp"
	var app models.App
	err := r.db.GetContext(ctx, &app,
		`SELECT id, name, secret FROM apps WHERE id = $1`, appId,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.App{}, errs.WithKind(op, errs.NotFound, storage.ErrAppNotFound)
		}
		return models.App{}, errs.WithKind(op, errs.Internal, err)
	}
	return app, nil
}

func (r *AppRepository) GetAppByIDTx(ctx context.Context, tx *sqlx.Tx, id int) (models.App, error) {
	const op = "appRepository.GetAppByIDTx"
	var app models.App
	err := tx.GetContext(ctx, &app, `SELECT id, name, secret FROM apps WHERE id = $1`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.App{}, errs.WithKind(op, errs.NotFound, storage.ErrAppNotFound)
		}
		return models.App{}, errs.WithKind(op, errs.Internal, err)
	}
	return app, nil
}
