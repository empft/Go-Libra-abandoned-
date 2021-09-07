package session

import (
	"context"
	"strings"
	"time"

	"github.com/stevealexrs/Go-Libra/database/kv"
	"github.com/stevealexrs/Go-Libra/random"
)

const (
	hashSetKey = "veiA"
)

type DefUniqueProvider struct {
	store kv.ExpiringHStore
	ns    string
}

func NewDefUniqueProvider(store kv.ExpiringHStore, namespace string) *DefUniqueProvider {
	return &DefUniqueProvider{
		store: store,
		ns:    namespace,
	}
}

func (sp *DefUniqueProvider) makeKey(keys ...string) string {
	return sp.ns + ":" + strings.Join(keys, ":")
}

func (sp *DefUniqueProvider) refreshExpiration(ctx context.Context, token string, expiration time.Duration) error {
	_, err := sp.store.Expire(ctx, sp.makeKey(token), expiration)
	return err
}

// The session must expire
func (sp *DefUniqueProvider) Init(ctx context.Context, expiration time.Duration) (*Unique, error) {
	token, err := random.Token20Byte()
	if err != nil {
		return nil, err
	}

	item := kv.HItem{
		Key: hashSetKey,
		Value: hashSetKey,
	}
	_, err = sp.store.HSet(ctx, sp.makeKey(token), item)
	if err != nil {
		return nil, err
	}
	err = sp.refreshExpiration(ctx, token, expiration)
	return &Unique{Token: token}, err
}

func (sp *DefUniqueProvider) Read(ctx context.Context, ssid string, expiration time.Duration) (*Unique, error) {
	sess := NewUnique(ssid)

	res, err := sp.store.HGet(ctx, sp.makeKey(sess.Token), hashSetKey)
	if err != nil {
		return nil, err
	}

	if res != hashSetKey {
		return nil, err
	}
	err = sp.refreshExpiration(ctx, sess.Token, expiration)
	return sess, err
}

func (sp *DefUniqueProvider) Destroy(ctx context.Context, ssid string) error {
	_, err := sp.Read(ctx, ssid, -1)
	return err
}

func (sp *DefUniqueProvider) Get(ctx context.Context, id, key string) (string, error) {
	masterKey := sp.makeKey(id)
	return sp.store.HGet(ctx, masterKey, key)
}

func (sp *DefUniqueProvider) Set(ctx context.Context, id, key, value string) error {
	if key == hashSetKey {
		panic("disallowed key value")
	}
	masterKey := sp.makeKey(id)

	item := kv.HItem{
		Key: key,
		Value: value,
	}
	_, err := sp.store.HSet(ctx, masterKey, item)
	return err
}

// Delete all items stored in session storage 
func (sp *DefUniqueProvider) DeleteAll(ctx context.Context, id string) error {
	masterKey := sp.makeKey(id)
	res, err := sp.store.HKeys(ctx, masterKey)
	if err != nil {
		return err
	}
	
	fields := make([]string, 0)
	for _, v := range res {
		if v != hashSetKey {
			fields = append(fields, v)
		}
	}
	_, err = sp.store.HDel(ctx, masterKey, fields...)
	return err
}