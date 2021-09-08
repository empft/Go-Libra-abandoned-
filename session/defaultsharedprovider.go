package session

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/stevealexrs/Go-Libra/database/kv"
	"github.com/stevealexrs/Go-Libra/random"
)

// sorted set and hash set store
type zhStore interface {
	kv.ZStore
	kv.HStore
}

type DefSharedProvider struct {
	session kv.ZStore
	item kv.HStore
	ns string
}

func NewDefSharedProvider(store zhStore, namespace string) *DefSharedProvider {
	return &DefSharedProvider{
		session: store,
		item: store,
		ns: namespace,
	}
}

func NewDefSharedProviderWithZHStore(zStore kv.ZStore, hStore kv.HStore, namespace string) *DefSharedProvider {
	return &DefSharedProvider{
		session: zStore,
		item: hStore,
		ns: namespace,
	}
}

func (sp *DefSharedProvider) makeKey(keys ...string) string {
	return sp.ns + ":" + strings.Join(keys, ":")
}

func (sp *DefSharedProvider) idToTokenKey(id int) string {
	return sp.makeKey("session", strconv.Itoa(id))
}

func (sp *DefSharedProvider) idToStoreKey(id int) string {
	return sp.makeKey("storage", strconv.Itoa(id))
}

func (sp *DefSharedProvider) createOrUpdateSession(ctx context.Context, id int, token string, lastAccessed int64) error {
	key := sp.idToTokenKey(id)

	return sp.session.ZAdd(ctx, key, kv.ZItem{
		Score: float64(lastAccessed), 
		Member: token,
	})
}

func (sp *DefSharedProvider) fetchAllSessions(ctx context.Context, id int) ([]Shared, error) {
	key := sp.idToTokenKey(id)
	res, err := sp.session.ZRangeWithScores(ctx, key, 0, -1)
	if err != nil {
		return nil, err
	}

	session := make([]Shared, len(res))
	for i, v := range res {
		session[i].Id = id
		session[i].Token = v.Member
		session[i].LastAccessed = int64(v.Score)
	}
	return session, nil	
}

func (sp *DefSharedProvider) deleteSession(ctx context.Context, id int, tokens ...string) error {
	key := sp.idToTokenKey(id)
	_, err := sp.session.ZRem(ctx, key, tokens...)
	return err
}

func (sp *DefSharedProvider) deleteStore(ctx context.Context, id int) error {
	key := sp.idToStoreKey(id)
	_, err := sp.item.Delete(ctx, key)
	return err
}
 
func (sp *DefSharedProvider) Init(ctx context.Context, id int) (*Shared, error) {
	token, err := random.Token20Byte()
	if err != nil {
		return nil, err
	}

	lastAccessed := time.Now().Unix()
	err = sp.createOrUpdateSession(ctx, id, token, lastAccessed)
	if err != nil {
		return nil, err
	}

	return &Shared{
		Id: id, 
		Token: token,
		LastAccessed: lastAccessed,
	}, nil
}

// Accept serialized session id
func (sp *DefSharedProvider) Read(ctx context.Context, ssid string) (*Shared, error) {
	sharedSession, err := NewShared(ssid)
	if err != nil {
		return nil, err
	}

	allSessions, err := sp.fetchAllSessions(ctx, sharedSession.Id)
	if err != nil {
		return nil, err
	}

	for i := range allSessions {
		if allSessions[i].Token == sharedSession.Token {
			sp.createOrUpdateSession(ctx, sharedSession.Id, sharedSession.Token, sharedSession.LastAccessed)
			return sharedSession, nil
		}
	}
	return nil, errors.New("session does not exist")
}

func (sp *DefSharedProvider) FetchAll(ctx context.Context, ssid string) ([]Shared, error) {
	session, err := sp.Read(ctx, ssid)
	if err != nil {
		return nil, err
	}

	return sp.fetchAllSessions(ctx, session.Id)
}

func (sp *DefSharedProvider) Destroy(ctx context.Context, ssid string) error {
	session, err := sp.Read(ctx, ssid)
	if err != nil {
		return err
	}

	allSessions, err := sp.fetchAllSessions(ctx, session.Id)
	if err != nil {
		return err
	}

	if len(allSessions) == 1 && allSessions[0].Token == session.Token {
		err := sp.deleteStore(ctx, session.Id)
		if err != nil {
			return err
		}
	}
	return sp.deleteSession(ctx, session.Id, session.Token)
}

func (sp *DefSharedProvider) DestroyOther(ctx context.Context, ssid string) error {
	session, err := sp.Read(ctx, ssid)
	if err != nil {
		return err
	}

	allSessions, err := sp.fetchAllSessions(ctx, session.Id)
	if err != nil {
		return err
	}

	var tokens []string
	for _, v := range allSessions {
		if v.Token != session.Token {
			tokens = append(tokens, v.Token)
		}
	}
	return sp.deleteSession(ctx, session.Id, tokens...)
}

// Objects are stored as hash in redis
func (sp *DefSharedProvider) Get(ctx context.Context, id, key string) (string, error) {
	idString, err := strconv.Atoi(id)
	if err != nil {
		return "", err
	}
	masterKey := sp.idToStoreKey(idString)
	return sp.item.HGet(ctx, masterKey, key)
}

func (sp *DefSharedProvider) Set(ctx context.Context, id, key string, value string) error {
	idString, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	masterKey := sp.idToStoreKey(idString)
	item := kv.HItem{
		Key: key,
		Value: value,
	}
	_, err = sp.item.HSet(ctx, masterKey, item)
	return err
}

func (sp *DefSharedProvider) DeleteAll(ctx context.Context, id string) error {
	idString, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	masterKey := sp.idToStoreKey(idString)
	_, err = sp.item.Delete(ctx, masterKey)
	return err
}