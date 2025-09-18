package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
)

func New(logger *slog.Logger, postgresConnectionString string) (*pgx.Conn, error) {
	db, err := pgx.Connect(context.Background(), postgresConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect postgres: %w", err)
	}

	return db, nil
}

type UnitOfWork struct {
	db *pgx.Conn
}

func NewUnitOfWork(db *pgx.Conn) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := u.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
