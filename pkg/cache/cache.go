package cache

import (
	"fmt"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
	"time"
)

const (
	CommonDb = 0
)

type Cache struct {
	CommonRedis *redis.Client
}

// ConfigOptions includes redis config options.
type ConfigOptions struct {
	Addr   string
	Passwd string
}

func Setup(logger *zap.SugaredLogger, config *ConfigOptions) (cache *Cache) {
	cache = &Cache{
		CommonRedis: redis.NewClient(
			&redis.Options{
				Addr:        config.Addr,
				Password:    config.Passwd,
				DB:          CommonDb,
				IdleTimeout: 3 * time.Second,
			}),
	}
	if logger != nil {
		logger.Debug(fmt.Sprintf("%s cache init success!", config.Addr))
	}
	return
}
