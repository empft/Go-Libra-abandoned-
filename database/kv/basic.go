package kv

import (
	"context"	
	"time"
)

type Getter interface {
	Get(ctx context.Context, key string) (string, error)
}

type Exister interface {
	Exist(ctx context.Context, keys ...string) (int, error)
}

type Deleter interface {
	Delete(ctx context.Context, keys ...string) (int, error)
}

type Expirer interface {
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)
}

type Store interface {
	Set(ctx context.Context, key string, value interface{}) error
	Getter
	Deleter
	Exister
}

type ExpiringStore interface {
	SetWithExpiration(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Getter
	Deleter
	Exister
}