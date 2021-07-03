package connector

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stevealexrs/Go-Libra/entity/session"
	"github.com/stevealexrs/Go-Libra/framework"
	"github.com/stevealexrs/Go-Libra/random"
)

type RedisSharedSessionProvider struct {
	redis *framework.RedisHandler
	ns string
}

func NewRedisSharedSessionProvider(namespace string, handler *framework.RedisHandler) *RedisSharedSessionProvider {
	return &RedisSharedSessionProvider{
		redis: handler,
		ns: namespace,
	}
}

func (sp *RedisSharedSessionProvider) makeKey(keys ...string) string {
	return sp.ns + ":" + strings.Join(keys, ":")
}

func (sp *RedisSharedSessionProvider) idToTokenKey(id int) string {
	return sp.makeKey("session", strconv.Itoa(id))
}

func (sp *RedisSharedSessionProvider) idToStoreKey(id int) string {
	return sp.makeKey("store", strconv.Itoa(id))
}

func (sp *RedisSharedSessionProvider) createOrUpdateSession(id int, token string, lastAccessed int64) error {
	key := sp.idToTokenKey(id)

	return sp.redis.Conn.ZAdd(context.TODO(), key, &redis.Z{
		Score: float64(lastAccessed), 
		Member: token,
	}).Err()
}

func (sp *RedisSharedSessionProvider) fetchAllSessions(id int) ([]entity.SharedSession, error) {
	key := sp.idToTokenKey(id)
	res, err := sp.redis.Conn.ZRangeWithScores(context.TODO(), key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	session := make([]entity.SharedSession, len(res))
	for i, v := range res {
		session[i].Id = id
		session[i].Token = v.Member.(string)
		session[i].LastAccessed = int64(v.Score)
	}
	return session, nil	
}

func (sp *RedisSharedSessionProvider) deleteSession(id int, tokens ...string) error {
	key := sp.idToTokenKey(id)

	tokenInterface := make([]interface{}, len(tokens))
	for i, v := range tokens {
		tokenInterface[i] = v
	}

	return sp.redis.Conn.ZRem(context.TODO(), key, tokenInterface...).Err()
}

func (sp *RedisSharedSessionProvider) deleteStore(id int) error {
	key := sp.idToStoreKey(id)
	return sp.redis.Conn.Del(context.TODO(), key).Err()
}
 
func (sp *RedisSharedSessionProvider) SharedSessionInit(id int) (session *entity.SharedSession, err error) {
	token, err := random.Token20Byte()
	if err != nil {
		return nil, err
	}

	lastAccessed := time.Now().Unix()
	sp.createOrUpdateSession(id, token, lastAccessed)

	return &entity.SharedSession{
		Id: id, 
		Token: token,
		LastAccessed: lastAccessed,
	}, nil
}

// Accept serialized session id
func (sp *RedisSharedSessionProvider) SessionRead(ssid string) (session *entity.SharedSession, err error) {
	sharedSession, err := entity.NewSharedSession(ssid)
	if err != nil {
		return nil, err
	}

	allSessions, err := sp.fetchAllSessions(sharedSession.Id)
	if err != nil {
		return nil, err
	}

	for i := range allSessions {
		if allSessions[i].Token == sharedSession.Token {
			sp.createOrUpdateSession(sharedSession.Id, sharedSession.Token, sharedSession.LastAccessed)
			return sharedSession, nil
		}
	}
	return nil, errors.New("session does not exist")
}

func (sp *RedisSharedSessionProvider) SessionFetchAll(session *entity.SharedSession) ([]entity.SharedSession, error) {
	return sp.fetchAllSessions(session.Id)
}

func (sp *RedisSharedSessionProvider) SessionDestroy(session *entity.SharedSession) error {
	allSessions, err := sp.fetchAllSessions(session.Id)
	if err != nil {
		return err
	}

	if len(allSessions) == 1 && allSessions[0].Token == session.Token {
		err := sp.deleteStore(session.Id)
		if err != nil {
			return err
		}
	}
	return sp.deleteSession(session.Id, session.Token)
}

func (sp *RedisSharedSessionProvider) SharedSessionDestroyOther(session *entity.SharedSession) error {
	allSessions, err := sp.fetchAllSessions(session.Id)
	if err != nil {
		return err
	}

	var tokens []string
	for _, v := range allSessions {
		if v.Token != session.Token {
			tokens = append(tokens, v.Token)
		}
	}
	return sp.deleteSession(session.Id, tokens...)
}

// Objects are stored as hash in redis
func (sp *RedisSharedSessionProvider) Get(id int, key string) (string, error) {
	redisKey := sp.idToStoreKey(id)
	return sp.redis.Conn.HGet(context.TODO(), redisKey, key).Result()
}

func (sp *RedisSharedSessionProvider) Set(id int, key string, value string) error {
	redisKey := sp.idToStoreKey(id)
	return sp.redis.Conn.HSet(context.TODO(), redisKey, key, value).Err()
}