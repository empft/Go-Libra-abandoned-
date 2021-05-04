package framework

import (
	"context"
	"time")


type Cache interface {
	Store(ctx context.Context, key string, item interface{}, time time.Time) error
	Fetch(ctx context.Context, key string, item interface{}) error
	Exist(key string) (bool, error)
}

type KVStore interface {
	Store(key string, item interface{}) error
	Fetch(key string) (interface{}, error)
}



