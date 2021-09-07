package object

import (
	"context"
	"io"
)

type Getter interface {
	Get(context.Context, string) ([]byte, error)
}

type Setter interface {
	Set(context.Context, io.Reader) (string, error)
}

type Deleter interface {
	Delete(context.Context, ...string) error
}

// Convert file id to url
type URLFormatter interface {
	FormatURL(fid ...string) ([]string, error)
}

// Convert url to file id
type FIDFormatter interface {
	FormatFID(objectURL ...string) ([]string, error)
}

type Store interface {
	Getter
	Setter
	Deleter
	URLFormatter
	FIDFormatter
}
