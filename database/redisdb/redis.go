package redisdb

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stevealexrs/Go-Libra/database/kv"
)
type Handler struct {
	Client redis.UniversalClient
}

func NewRedisHandlerWithClient(client redis.UniversalClient) *Handler {
	return &Handler{Client: client}
}

func (handler *Handler) Set(ctx context.Context, key string, item interface{}) error {
	err := handler.Client.Set(
		ctx,
		key,
		item,
		0,
	).Err()

	return err
}

func (handler *Handler) SetWithExpiration(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := handler.Client.Set(
		ctx,
		key,
		value,
		expiration,
	).Err()

	return err
}

func (handler *Handler) Get(ctx context.Context, key string) (string, error) {
	item, err := handler.Client.Get(
		ctx,
		key,
	).Result()

	return item, err
}

func (handler *Handler) Exist(ctx context.Context, keys ...string) (int, error) {
	res, err := handler.Client.Exists(
		ctx,
		keys...,
	).Result()

	return int(res), err
}
func (handler *Handler) Delete(ctx context.Context, keys ...string) (int, error) {
	res, err := handler.Client.Del(ctx, keys...).Result()
	return int(res), err
}

func (handler *Handler) ZAdd(ctx context.Context, key string, values ...kv.ZItem) error {
	members := make([]*redis.Z, len(values))
	
	for i, v := range values {
		members[i] = &redis.Z{Score: v.Score, Member: v.Member}
	}

	return handler.Client.ZAdd(
		ctx,
		key,
		members...
	).Err()
}

func (handler *Handler) ZRangeWithScores(ctx context.Context, key string, start int, stop int) ([]kv.ZItem, error) {
	res, err := handler.Client.ZRangeWithScores(ctx, key, int64(start), int64(stop)).Result()
	items := make([]kv.ZItem, len(res))
	for i, v := range res {
		items[i] = kv.ZItem{Score: v.Score, Member: v.Member.(string)}
	}
	return items, err
}

func (handler *Handler) ZRem(ctx context.Context, key string, members ...string) (int, error) {
	memberInterface := make([]interface{}, len(members))
	for i, v := range members {
		memberInterface[i] = v
	}
	res, err := handler.Client.ZRem(ctx, key, memberInterface...).Result()
	return int(res), err
}

func (handler *Handler) HSet(ctx context.Context, key string, values ...kv.HItem) (int, error) {
	var valueInterface []interface{}
	for _, v := range values {
		valueInterface = append(valueInterface, v.Key, v.Value)
	}
	res, err := handler.Client.HSet(ctx, key, valueInterface...).Result()
	return int(res), err
}

func (handler *Handler) HGet(ctx context.Context, key, field string) (string, error) {
	return handler.Client.HGet(ctx, key, field).Result()
}

func (handler *Handler) HKeys(ctx context.Context, key string) ([]string, error) {
	return handler.Client.HKeys(ctx, key).Result()
}

func (handler *Handler) HDel(ctx context.Context, key string, fields ...string) (int, error) {
	res, err := handler.Client.HDel(ctx, key, fields...).Result()
	return int(res), err
}

func (handler *Handler) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return handler.Client.Expire(ctx, key, expiration).Result()
}