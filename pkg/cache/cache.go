package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ossm-org/orchid/pkg/logging"
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
func New(ctx context.Context, config *ConfigOptions) Cache {
	logger := logging.FromContext(ctx)

	opts := &redis.Options{
		Addr:     config.Addr,
		Password: config.Passwd,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			if _, err := cn.Ping(ctx).Result(); err != nil {
				return err
			}
			logger.Debugf("Successfully connected to Redis: %s", config.Addr)
			return nil
		},
		DB:          commonDb,
		MaxRetries:  maxRetry,
		IdleTimeout: idleTimeOut,
	}

	return Cache{
		Client: redis.NewClient(opts),
	}
}
