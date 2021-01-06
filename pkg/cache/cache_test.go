package cache

import (
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestCase struct {
	Exist  int64
	Hello  string
	Length int64
}

func TestCache(t *testing.T) {
	if err := Setup(); err != nil {
		t.Fatal(err)
	}
	testCase := &TestCase{
		Exist:  0,
		Hello:  "hello",
		Length: 2,
	}
	Storage.CommonRedis.Del(testKey)
	exist := Storage.CommonRedis.Exists(testKey).Val()
	assert.Equal(t, exist, testCase.Exist)
	Storage.CommonRedis.Set(testKey, "hello", 0)
	bytes, err := Storage.CommonRedis.Get(testKey).Bytes()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, string(bytes), testCase.Hello)
	Storage.CommonRedis.Del(testKey)
	exist = Storage.CommonRedis.Exists(testKey).Val()
	assert.Equal(t, exist, testCase.Exist)
	Storage.CommonRedis.ZAdd(testKey, redis.Z{Member: 1, Score: 1}, redis.Z{Member: 2, Score: 2})
	length := Storage.CommonRedis.ZCard(testKey).Val()
	assert.Equal(t, length, testCase.Length)
}
