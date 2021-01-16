package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Migrate executes database migrations.
func (db Database) Migrate(ctx context.Context) error {
	conn, err := db.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if err := createSchemaVersionTable(ctx, conn); err != nil {
		return err
	}

	var currentVersion int
	if err := conn.QueryRow(ctx, `SELECT version FROM schema_version`).Scan(&currentVersion); err != nil && err != pgx.ErrNoRows {
		return err
	}

	db.logger.Debug("-> Current schema version:", currentVersion)
	db.logger.Debug("-> Latest schema version:", schemaVersion)

	for version := currentVersion; version < schemaVersion; version++ {
		newVersion := version + 1
		db.logger.Debug("* Migrating to version:", newVersion)

		tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
		if err != nil {
			return fmt.Errorf("[Migration v%d] %v", newVersion, err)
		}

		if err := migrations[version](ctx, tx); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("[Migration v%d] %v", newVersion, err)
		}

		if _, err := tx.Exec(ctx, `DELETE FROM schema_version`); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("[Migration v%d] %v", newVersion, err)
		}

		if _, err := tx.Exec(ctx, `INSERT INTO schema_version (version) VALUES ($1)`, newVersion); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("[Migration v%d] %v", newVersion, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("[Migration v%d] %v", newVersion, err)
		}
	}

	return nil
}

// IsSchemaUpToDate checks if the database schema is up to date.
func (db Database) IsSchemaUpToDate(ctx context.Context) error {
	conn, err := db.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	var currentVersion int
	if err := conn.QueryRow(ctx, `SELECT version FROM schema_version`).Scan(&currentVersion); err != nil {
		return err
	}
	if currentVersion < schemaVersion {
		return fmt.Errorf(`the database schema is not up to date: current=v%d expected=v%d`, currentVersion, schemaVersion)
	}
	db.logger.Debug("Database schema is up to date")
	return nil
}

func createSchemaVersionTable(ctx context.Context, conn *pgxpool.Conn) error {
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_version (
			version integer not null
		);
	`)
	return err
}
