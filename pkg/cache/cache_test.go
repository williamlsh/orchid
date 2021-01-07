package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestCase struct {
	Exist int64
}

func TestCache(t *testing.T) {
	config := &ConfigOptions{
		Addr:   "localhost:6379",
		Passwd: "",
	}
	cache := Setup(nil, config)
	testKey := "test:key"
	testCase := &TestCase{
		Exist: 0,
	}
	cache.CommonRedis.Del(testKey)
	exist := cache.CommonRedis.Exists(testKey).Val()
	assert.Equal(t, exist, testCase.Exist)
}
