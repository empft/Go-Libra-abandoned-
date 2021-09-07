package session

import "context"

type Storage interface {
	Get(ctx context.Context, id, key string) (string, error)
	Set(ctx context.Context, id, key, value string) error
	DeleteAll(ctx context.Context, id string) error
}