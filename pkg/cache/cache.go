package cache

import (
	"github.com/go-redis/redis"
	"time"
)

const (
	CommonDb = 0
)

var (
	Storage *Cache
)

type Cache struct {
	CommonRedis *redis.Client
}

func Setup() error {
	Storage = &Cache{}
	Storage.CommonRedis = redis.NewClient(
		&redis.Options{
			//fixme move to config center
			Addr:        "localhost:6379",
			Password:    "",
			DB:          CommonDb,
			IdleTimeout: 3 * time.Second,
		})

	return nil
}
