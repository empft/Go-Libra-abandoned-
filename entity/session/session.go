package entity

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type SharedSessionProvider interface {
	SharedSessionInit(id int) (*SharedSession, error)
	SharedSessionDestroyOther(*SharedSession) error
	SessionRead(string) (*SharedSession, error)
	SessionFetchAll(*SharedSession) ([]SharedSession, error)
	SessionDestroy(*SharedSession) error
	Get(id int, key string) (string, error)
	Set(id int, key, value string) error
}

type UniqueSessionProvider interface {
	SessionInit(expiration time.Duration) (*UniqueSession, error)
	SessionRead(string, time.Duration) (*UniqueSession, error)
	SessionDestroy(*UniqueSession) error
	Get(token, key string) (string, error)
	Set(token, key, value string) error
}

type SharedSession struct {
	Id int
	Token string
	LastAccessed int64
}

func NewSharedSession(ssid string) (*SharedSession, error) {
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
	return &SharedSession{Id: id, Token: token, LastAccessed: lastAccessed}, nil
}

func (s *SharedSession) SessionId() string {
	return strconv.Itoa(s.Id) + "~" + s.Token
}

type UniqueSession struct {
	Token string
}

func NewUniqueSession(ssid string) *UniqueSession {
	return &UniqueSession{ssid}
}

func (s *UniqueSession) SessionId() string {
	return s.Token
}