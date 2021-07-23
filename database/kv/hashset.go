package kv

import "context"

type HStore interface {
	HSet(ctx context.Context, key string, values ...HItem) (int, error)
	HGet(ctx context.Context, key, field string) (string, error)
	HKeys(ctx context.Context, key string) ([]string, error)
	HDel(ctx context.Context, key string, fields ...string) (int, error)
	Deleter
}

type ExpiringHStore interface {
	HStore
	Expirer
}

type HItem struct {
	Key string
	Value string
}