package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stevealexrs/Go-Libra/database/kv"
)
type RedisHandler struct {
	Client redis.UniversalClient
}

func NewRedisHandler(password string) *RedisHandler {
	return &RedisHandler{redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName: "redis-master",
		SentinelAddrs: []string{":9126", ":9127", ":9128"},
		Password: password,
	})}
}

func NewRedisHandlerWithClient(client redis.UniversalClient) *RedisHandler {
	return &RedisHandler{Client: client}
}

func (handler *RedisHandler) Set(ctx context.Context, key string, item interface{}) error {
	err := handler.Client.Set(
		ctx,
		key,
		item,
		0,
	).Err()

	return err
}

func (handler *RedisHandler) SetWithExpiration(ctx context.Context, key string, item interface{}, expiration time.Duration) error {
	err := handler.Client.Set(
		ctx,
		key,
		item,
		expiration,
	).Err()

	return err
}

func (handler *RedisHandler) Get(ctx context.Context, key string) (string, error) {
	item, err := handler.Client.Get(
		ctx,
		key,
	).Result()

	return item, err
}

func (handler *RedisHandler) Exists(ctx context.Context, key ...string) (int, error) {
	res, err := handler.Client.Exists(
		ctx,
		key...,
	).Result()

	return int(res), err
}
func (handler *RedisHandler) Delete(ctx context.Context, keys ...string) (int, error) {
	res, err := handler.Client.Del(ctx, keys...).Result()
	return int(res), err
}

func (handler *RedisHandler) ZAdd(ctx context.Context, key string, values ...kv.ZItem) error {
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

func (handler *RedisHandler) ZRangeWithScores(ctx context.Context, key string, start int, stop int) ([]kv.ZItem, error) {
	res, err := handler.Client.ZRangeWithScores(ctx, key, int64(start), int64(stop)).Result()
	items := make([]kv.ZItem, len(res))
	for i, v := range res {
		items[i] = kv.ZItem{Score: v.Score, Member: v.Member.(string)}
	}
	return items, err
}

func (handler *RedisHandler) ZRem(ctx context.Context, key string, members ...string) (int, error) {
	memberInterface := make([]interface{}, len(members))
	for i, v := range members {
		memberInterface[i] = v
	}
	res, err := handler.Client.ZRem(ctx, key, memberInterface...).Result()
	return int(res), err
}

func (handler *RedisHandler) HSet(ctx context.Context, key string, values ...kv.HItem) (int, error) {
	var valueInterface []interface{}
	for _, v := range values {
		valueInterface = append(valueInterface, v.Key, v.Value)
	}
	res, err := handler.Client.HSet(ctx, key, valueInterface...).Result()
	return int(res), err
}

func (handler *RedisHandler) HGet(ctx context.Context, key, field string) (string, error) {
	return handler.Client.HGet(ctx, key, field).Result()
}

func (handler *RedisHandler) HKeys(ctx context.Context, key string) ([]string, error) {
	return handler.Client.HKeys(ctx, key).Result()
}

func (handler *RedisHandler) HDel(ctx context.Context, key string, fields ...string) (int, error) {
	res, err := handler.Client.HDel(ctx, key, fields...).Result()
	return int(res), err
}

func (handler *RedisHandler) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return handler.Client.Expire(ctx, key, expiration).Result()
}