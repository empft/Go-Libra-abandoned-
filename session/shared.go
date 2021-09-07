package session

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
)


type SharedProvider interface {
	Init(ctx context.Context, id int) (*Shared, error)
	Read(ctx context.Context, ssid string) (*Shared, error)
	FetchAll(ctx context.Context, ssid string) ([]Shared, error)
	Destroy(ctx context.Context, ssid string) error
	DestroyOther(ctx context.Context, ssid string) error
	Storage
}

type Shared struct {
	Id           int
	Token        string
	LastAccessed int64
}

func NewShared(ssid string) (*Shared, error) {
	ss := strings.Split(ssid, "~")
	if len(ss) != 2 {
		return nil, errors.New("invalid session id")
	}

	id, err := strconv.Atoi(ss[0])
	if err != nil {
		return nil, err
	}
	token := ss[1]

	lastAccessed := time.Now().Unix()
	return &Shared{Id: id, Token: token, LastAccessed: lastAccessed}, nil
}

func (s *Shared) SessionId() string {
	return strconv.Itoa(s.Id) + "~" + s.Token
}