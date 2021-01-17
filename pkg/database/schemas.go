package database

import (
	"context"

	"github.com/jackc/pgx/v4"
)

var schemaVersion = len(migrations)

// Add schemas here whenever you update database schemas.
var migrations = []func(ctx context.Context, tx pgx.Tx) error{
	func(ctx context.Context, tx pgx.Tx) (err error) {
		sql := `
			CREATE TABLE IF NOT EXISTS users(
				id serial PRIMARY KEY,
				username VARCHAR (50) UNIQUE NOT NULL,
				email VARCHAR (300) UNIQUE NOT NULL
		 	);
		`
		_, err = tx.Exec(ctx, sql)
		return err
	},
	func(ctx context.Context, tx pgx.Tx) (err error) {
		sql := `
			ALTER TABLE users
			ADD COLUMN alias VARCHAR (50),
			ADD COLUMN deregistered boolean DEFAULT false;
		`
		_, err = tx.Exec(ctx, sql)
		return err
	},
	func(ctx context.Context, tx pgx.Tx) (err error) {
		sql := `
			ALTER TABLE users
			ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

			CREATE OR REPLACE FUNCTION trigger_set_timestamp()
			RETURNS TRIGGER AS $$
			BEGIN
  				NEW.updated_at = NOW();
  				RETURN NEW;
			END;
			$$ LANGUAGE plpgsql;

			CREATE TRIGGER set_timestamp
			BEFORE UPDATE ON users
			FOR EACH ROW
			EXECUTE PROCEDURE trigger_set_timestamp();
		`
		_, err = tx.Exec(ctx, sql)
		return err
	},
}
