package framework

import (
	"time"
	"context"

	"github.com/go-redis/redis/v8"
    "github.com/go-redis/cache/v8"
)

type RedisHandler struct {
	Conn *redis.Client
}

func NewRedisHandler(password string) *RedisHandler {
	return &RedisHandler{redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName: "redis-master",
		SentinelAddrs: []string{":9126", ":9127", ":9128"},
		Password: password,
	})}
}

func (handler *RedisHandler) Store(ctx context.Context, key string, item interface{}) error {
	err := handler.Conn.Set(
		ctx,
		key,
		item,
		0,
	).Err()

	return err
}

func (handler *RedisHandler) Fetch(ctx context.Context, key string) (interface{}, error) {
	item, err := handler.Conn.Get(
		ctx,
		key,
	).Result()

	return item, err
}

type RedisCacheHandler struct {
	Conn *cache.Cache
}

func NewRedisCacheHandler(handler *RedisHandler) *RedisCacheHandler {
	return &RedisCacheHandler{cache.New(&cache.Options{
		Redis: handler.Conn,
		LocalCache: cache.NewTinyLFU(10000, time.Minute), // Tiny cache in current server
	})}
}

func (handler *RedisCacheHandler) Store(ctx context.Context, key string, item interface{}, exp time.Duration) error {
	err := handler.Conn.Set(&cache.Item{
        Ctx:   ctx,
        Key:   key,
        Value: item,
        TTL:   exp,
    })

	return err
}

func (handler *RedisCacheHandler) Fetch(ctx context.Context, key string, item interface{}) error {
	err := handler.Conn.Get(
		ctx,
		key,
		&item,
	)

	return err
}

func (handler *RedisCacheHandler) Exist(ctx context.Context, key string) bool {
	return handler.Conn.Exists(
		ctx,
		key,
	)
}




