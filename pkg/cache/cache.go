package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	commonDb    = 0
	maxRetry    = 5
	idleTimeOut = 5 * time.Minute
)

// Cache is a redis cache.
type Cache struct {
	Client *redis.Client
}

// ConfigOptions includes redis config options.
type ConfigOptions struct {
	Addr   string
	Passwd string
}

// New returns a new redis cache.
func New(logger *zap.SugaredLogger, config *ConfigOptions) Cache {
	opts := &redis.Options{
		Addr:        config.Addr,
		Password:    config.Passwd,
		DB:          commonDb,
		MaxRetries:  maxRetry,
		IdleTimeout: idleTimeOut,
	}

	client := redis.NewClient(opts)
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		panic(err)
	}
	logger.Debugf("Successfully connected to redis, address=%s", config.Addr)

	return Cache{
		Client: client,
	}
}
