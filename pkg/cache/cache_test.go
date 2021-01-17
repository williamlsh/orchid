package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type TestCase struct {
	Exist int64
}

func TestCache(t *testing.T) {
	config := &ConfigOptions{
		Addr:   "localhost:6379",
		Passwd: "",
	}
	cache := New(zap.NewExample().Sugar(), config)
	testKey := "test:key"
	testCase := &TestCase{
		Exist: 0,
	}
	cache.Client.Del(context.Background(), testKey)
	exist := cache.Client.Exists(context.Background(), testKey).Val()
	assert.Equal(t, exist, testCase.Exist)
}
