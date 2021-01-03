package cache

import (
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

// Cache fullfills redis caching capabilities.
type Cache interface {
	Set(args ...interface{}) error
	Get(args ...interface{}) (interface{}, error)
	Delete(args ...interface{}) (int, error)
}

type cache struct {
	logger *zap.SugaredLogger
	pool   *redis.Pool
}

// New returns a new cache.
func New(addr, passwd string, logger *zap.SugaredLogger) Cache {
	return cache{
		logger: logger,
		pool:   newPool(addr, passwd),
	}
}

func (c cache) Set(args ...interface{}) error {
	conn := c.pool.Get()
	defer conn.Close()

	_, err := redis.String(conn.Do("SET", args...))
	return err
}

func (c cache) Get(args ...interface{}) (interface{}, error) {
	conn := c.pool.Get()
	defer conn.Close()

	reply, err := conn.Do("GET", args...)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (c cache) Delete(args ...interface{}) (int, error) {
	conn := c.pool.Get()
	defer conn.Close()

	return redis.Int(conn.Do("DEL", args...))
}
