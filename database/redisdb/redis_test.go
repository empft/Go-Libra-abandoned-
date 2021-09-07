package redisdb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stevealexrs/Go-Libra/database/redisdb"
)

var ctx = context.TODO()

func SampleDataForRedis(handler *redisdb.Handler, newsID int) (string, error) {
	cacheKey := fmt.Sprintf("news_redis_cache_%d", newsID)

	info, err := handler.Get(ctx, cacheKey)
	if err == redis.Nil {
		// info, err = call api()
		info = "test"
		err = handler.Set(ctx, cacheKey, info)
	}

	return info, err
}

func SampleDataForCache(handler *redisdb.RedisCacheHandler, newsID int) (string, error) {
	cacheKey := fmt.Sprintf("news_redis_cache_%d", newsID)

	info, err := handler.Get(ctx, cacheKey)
	if err == cache.ErrCacheMiss {
		// info, err = call api()
		info = "test"
		err = handler.Set(ctx, cacheKey, info, 30 * time.Minute)
	}

	return info.(string), err
}

func TestRedisClient(t *testing.T) {
	db, mock := redismock.NewClientMock()

	// Create the redis handler directly without default constructor
	handler := &redisdb.Handler{db}

	newsID := 123456789
	key := fmt.Sprintf("news_redis_cache_%d", newsID)

	// mock ignoring `call api()`

	mock.ExpectGet(key).RedisNil()
	mock.Regexp().ExpectSet(key, `[a-z]+`, 0).SetErr(errors.New("FAIL"))

	_, err := SampleDataForRedis(handler, newsID)
	if err == nil || err.Error() != "FAIL" {
		t.Error("wrong error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRedisCacheClient(t *testing.T) {
	db, mock := redismock.NewClientMock()

	// Create the redis handler directly without default constructor
	handler := &redisdb.Handler{db}
	cacheHandler := redisdb.NewRedisCacheHandler(handler)

	newsID := 123
	key := fmt.Sprintf("news_redis_cache_%d", newsID)

	// mock ignoring `call api()`

	mock.ExpectGet(key).SetErr(cache.ErrCacheMiss)
	// will work if set expected value from [a-z]+ to any. The value set becomes decimal representation of ascii.
	mock.Regexp().ExpectSet(key, `[a-z]+`, 30 * time.Minute).SetErr(errors.New("FAIL"))

	_, err := SampleDataForCache(cacheHandler, newsID)
	if err == nil || err.Error() != "FAIL" {
		t.Error("wrong error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}