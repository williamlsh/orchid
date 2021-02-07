package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/williamlsh/orchid/pkg/logging"
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
func New(ctx context.Context, config *ConfigOptions) Cache {
	logger := logging.FromContext(ctx)
	redis.SetLogger(redisLoggerAdapter{logger})

	opts := &redis.Options{
		Addr:        config.Addr,
		Password:    config.Passwd,
		DB:          commonDb,
		MaxRetries:  maxRetry,
		IdleTimeout: idleTimeOut,
	}

	client := redis.NewClient(opts)
	if _, err := client.Ping(ctx).Result(); err != nil {
		logger.Panic(err)
	}
	logger.Debugf("Successfully connected to Redis: %s", config.Addr)

	return Cache{client}
}

// Compile time interface check.
var _ interface {
	Printf(ctx context.Context, format string, v ...interface{})
} = redisLoggerAdapter{}

// redisLoggerAdapter implements redis internal Logging interface.
type redisLoggerAdapter struct {
	logger *zap.SugaredLogger
}

func (l redisLoggerAdapter) Printf(_ context.Context, format string, v ...interface{}) {
	l.logger.Infof(format, v)
}
