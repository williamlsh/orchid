package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

// Database represents a new database.
type Database struct {
	Pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

// New returns a new database.
func New(logger *zap.SugaredLogger, dsn string) Database {
	return Database{
		Pool:   newPool(logger, dsn),
		logger: logger,
	}
}

// newPool returns a new postgres connection pool.
func newPool(logger *zap.SugaredLogger, dsn string) *pgxpool.Pool {
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}

	poolConfig.ConnConfig.Logger = zapadapter.NewLogger(logger.Desugar())

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		panic(err)
	}

	return pool
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
