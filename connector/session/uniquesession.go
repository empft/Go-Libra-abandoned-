package connector

import (
	"context"
	"strings"
	"time"

	"github.com/stevealexrs/Go-Libra/entity/session"
	"github.com/stevealexrs/Go-Libra/framework"
	"github.com/stevealexrs/Go-Libra/random"
)

const (
	hashSetKey = "veiA"
)

type RedisSessionProvider struct {
	redis *framework.RedisHandler
	ns    string
}

func NewRedisSessionProvider(namespace string, handler *framework.RedisHandler) *RedisSessionProvider {
	return &RedisSessionProvider{
		redis: handler,
		ns:    namespace,
	}
}

func (sp *RedisSessionProvider) makeKey(keys ...string) string {
	return sp.ns + ":" + strings.Join(keys, ":")
}

func (sp *RedisSessionProvider) sessionRefresh(token string, expiration time.Duration) error {
	return sp.redis.Conn.Expire(context.TODO(), sp.makeKey(token), expiration).Err()
}

func (sp *RedisSessionProvider) SessionInit(expiration time.Duration) (*entity.UniqueSession, error) {
	token, err := random.Token20Byte()
	if err != nil {
		return nil, err
	}

	err = sp.redis.Conn.HSet(context.TODO(), sp.makeKey(token), hashSetKey, hashSetKey).Err()
	if err != nil {
		return nil, err
	}
	err = sp.sessionRefresh(token, expiration)
	return &entity.UniqueSession{Token: token}, err
}

// Server should generate a new SessionId for initializing a session
func (sp *RedisSessionProvider) SessionRead(ssid string, expiration time.Duration) (*entity.UniqueSession, error) {
	session := entity.NewUniqueSession(ssid)

	res, err := sp.redis.Conn.HGet(context.TODO(), sp.makeKey(session.Token), hashSetKey).Result()
	if err != nil {
		return nil, err
	}

	if res != hashSetKey {
		return nil, err
	}
	err = sp.sessionRefresh(session.Token, expiration)
	return session, err
}

func (sp *RedisSessionProvider) SessionDestroy(session entity.UniqueSession) error {
	return sp.redis.Conn.Del(context.TODO(), sp.makeKey(session.Token)).Err()
}

// Clear all other sessions that belong to the same user
func (sp *RedisSessionProvider) Get(token, key string) (string, error) {
	redisKey := sp.makeKey(token)
	return sp.redis.Conn.HGet(context.TODO(), redisKey, key).Result()
}

func (sp *RedisSessionProvider) Set(token, key string, value string) error {
	if key == hashSetKey {
		panic("disallowed key value")
	}
	redisKey := sp.makeKey(token)
	return sp.redis.Conn.HSet(context.TODO(), redisKey, key, value).Err()
}