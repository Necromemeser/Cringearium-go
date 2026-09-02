package db

import (
	"context"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {

	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

// TODO: вписать названия таблиц в команде дропа
func (db *DB) Drop(ctx context.Context) error {
	db.log.Info("starting Drop")

	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		db.log.Error("failed to begin transaction", "error", err)
		return err
	}

	defer func() {
		if err != nil {
			db.log.Warn("rolling back transaction", "error", err)
			_ = tx.Rollback()
		}
	}()

	db.log.Debug("dropping comic_keywords table")
	_, err = tx.ExecContext(ctx, `
    	DROP TABLE *НАЗВАНИЯ ТАБЛИЦ*
	`)
	if err != nil {
		db.log.Error("failed to drop db", "error", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		db.log.Error("failed to commit transaction", "error", err)
		return err
	}

	db.log.Info("drop completed successfully")
	return nil
}
