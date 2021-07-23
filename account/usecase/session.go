package account

import (
	"github.com/stevealexrs/Go-Libra/session/entity"
)

type Session struct {
	session.Shared
	provider session.SharedProvider
}

func NewSession(baseSession session.Shared, provider session.SharedProvider) *Session {
	return &Session{
		Shared: baseSession,
		provider: provider,
	}
}