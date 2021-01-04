package database

import (
	"context"

	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

// Database represents a new database.
type Database struct {
	Pool *pgxpool.Pool
}

// New returns a new database.
func New(logger *zap.SugaredLogger, dsn string) Database {
	return Database{
		Pool: newPool(logger, dsn),
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
