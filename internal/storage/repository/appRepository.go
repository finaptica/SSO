package repository

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/finaptica/sso/internal/domain/models"
	"github.com/finaptica/sso/internal/lib/errs"
	"github.com/jackc/pgx/v5"
)

type AppRepository struct {
	db  *pgx.Conn
	log *slog.Logger
}

func NewAppRepository(log *slog.Logger, db *pgx.Conn) *AppRepository {
	return &AppRepository{log: log, db: db}
}

func (r *AppRepository) GetAppById(ctx context.Context, appId int) (models.App, error) {
	const op = "appRepository.GetApp"
	var app models.App
	err := r.db.QueryRow(ctx, "SELECT id, name, secret FROM apps WHERE id = $1", appId).Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.App{}, errs.WithKind(op, errs.NotFound, err)
		}
		return models.App{}, errs.WithKind(op, errs.Internal, err)
	}
	return app, nil
}

func (r *AppRepository) GetAppByIDTx(ctx context.Context, tx pgx.Tx, id int) (models.App, error) {
	const op = "appRepository.GetAppByIDTx"
	var app models.App
	err := tx.QueryRow(ctx, "SELECT id,name,secret FROM apps WHERE id = $1", id).Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.App{}, errs.WithKind(op, errs.NotFound, err)
		}
		return models.App{}, errs.WithKind(op, errs.Internal, err)
	}
	return app, nil
}
