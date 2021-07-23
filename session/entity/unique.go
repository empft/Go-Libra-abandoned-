package session

import "time"

type UniqueProvider interface {
	Init(expiration time.Duration) (*Unique, error)
	Read(string, time.Duration) (*Unique, error)
	Destroy(string) error
	Storage
}

type Unique struct {
	Token string
}

func NewUnique(ssid string) *Unique {
	return &Unique{ssid}
}

func (s *Unique) SessionId() string {
	return s.Token
}