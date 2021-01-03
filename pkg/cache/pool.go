package cache

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Redis dial options.
const (
	dialConnectTimeout = 3 * time.Second
	dialReadTimeout    = 3 * time.Second
	dialWriteTimeout   = 3 * time.Second
)

func newPool(addr, passwd string) *redis.Pool {
	dialOpts := []redis.DialOption{
		redis.DialPassword(passwd),
		redis.DialConnectTimeout(dialConnectTimeout),
		redis.DialReadTimeout(dialReadTimeout),
		redis.DialWriteTimeout(dialWriteTimeout),
	}
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", addr, dialOpts...)
			if err != nil {
				return nil, fmt.Errorf("could not dial: %v", err)
			}
			return conn, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}
