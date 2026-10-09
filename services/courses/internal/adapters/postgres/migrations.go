package postgres

import (
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql migrations/legacy/*.sql
var migrationFiles embed.FS

func (db *DB) Migrate() error {
	db.log.Debug("running migration")

	// Existing databases have a schema_migrations table but no squashed-baseline
	// marker. Keep their original migration history to avoid reinterpreting versions.
	var useLegacy bool
	err := db.conn.QueryRow(`SELECT
		to_regclass('public.schema_migrations') IS NOT NULL
		AND to_regclass('public.course_migration_profile') IS NULL`).Scan(&useLegacy)
	if err != nil {
		return err
	}

	migrationDir := "migrations"
	if useLegacy {
		migrationDir = "migrations/legacy"
	}

	files, err := iofs.New(migrationFiles, migrationDir)
	if err != nil {
		return err
	}

	driver, err := pgx.WithInstance(db.conn.DB, &pgx.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		files,
		"pgx",
		driver,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil {
		if err != migrate.ErrNoChange {
			db.log.Error("migration failed", "error", err)
			return err
		}

		db.log.Debug("migration did not change anything")
	}

	db.log.Debug("migration finished")
	return nil
}
