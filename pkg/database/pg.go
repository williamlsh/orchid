package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ossm-org/orchid/pkg/logging"
	"go.uber.org/zap"
)

// Database represents a new database.
type Database struct {
	Pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

// New returns a new database.
func New(ctx context.Context, dsn string) Database {
	logger := logging.FromContext(ctx)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}

	poolConfig.ConnConfig.Logger = zapadapter.NewLogger(logger.Desugar())

	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		panic(err)
	}

	return Database{
		Pool:   pool,
		logger: logger,
	}
}

// InTx runs the given function f within a transaction with isolation level serialization by default.
func (db Database) InTx(ctx context.Context, fn ...func(tx pgx.Tx) error) error {
	conn, err := db.Pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquiring connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}

	for _, f := range fn {
		if err := f(tx); err != nil {
			if err1 := tx.Rollback(ctx); err1 != nil {
				return fmt.Errorf("rolling back transaction: %w (original error: %v)", err1, err)
			}
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %v", err)
	}
	return nil
}
