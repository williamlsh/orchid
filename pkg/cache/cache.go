package cache

import (
	"net"
	"time"

	"github.com/go-redis/redis"
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
		Dialer: func() (net.Conn, error) {
			return net.Dial("tcp", config.Addr)
		},
		OnConnect: func(c *redis.Conn) error {
			if _, err := c.Ping().Result(); err != nil {
				return err
			}
			logger.Debugf("Successfully connected to Redis: %s\n", config.Addr)
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
