package kv

import "context"

// Sorted Set store
type ZStore interface {
	ZAdd(ctx context.Context, key string, values ...ZItem) error
	ZRangeWithScores(ctx context.Context, key string, start int, stop int) ([]ZItem, error)
	ZRem(ctx context.Context, key string, members ...string) error
	Deleter
}

type ZItem struct {
	Score float64
	Member  string
}