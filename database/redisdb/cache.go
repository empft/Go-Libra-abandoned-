package redisdb

import (
	"context"
	"time"

	"github.com/go-redis/cache/v8"
)


type RedisCacheHandler struct {
	Conn *cache.Cache
}

func NewRedisCacheHandler(handler *Handler) *RedisCacheHandler {
	return &RedisCacheHandler{cache.New(&cache.Options{
		Redis:      handler.Client,
		LocalCache: cache.NewTinyLFU(10000, time.Minute), // Tiny cache in current server
	})}
}

func (handler *RedisCacheHandler) Set(ctx context.Context, key string, item interface{}, exp time.Duration) error {
	err := handler.Conn.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: item,
		TTL:   exp,
	})

	return err
}

func (handler *RedisCacheHandler) Get(ctx context.Context, key string) (interface{}, error) {
	var item interface{}
	err := handler.Conn.Get(
		ctx,
		key,
		&item,
	)

	return item, err
}

func (handler *RedisCacheHandler) Exist(ctx context.Context, key string) bool {
	return handler.Conn.Exists(
		ctx,
		key,
	)
}