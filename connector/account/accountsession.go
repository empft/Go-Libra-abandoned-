package connector

import (
	"github.com/stevealexrs/Go-Libra/connector/session"
	"github.com/stevealexrs/Go-Libra/entity/session"
	"github.com/stevealexrs/Go-Libra/framework"
	"github.com/stevealexrs/Go-Libra/namespace"
)

type AccountSessionProvider struct {
	provider entity.SharedSessionProvider
}

func NewAccountSessionProviderFromRedis(handler *framework.RedisHandler) *AccountSessionProvider {
	return &AccountSessionProvider{
		provider: connector.NewRedisSharedSessionProvider(namespace.RedisAccSharedSession, handler),
	}
}

func (sp *AccountSessionProvider) SessionInit(accountId int) (*AccountSession, error) {
	session, err := sp.provider.SharedSessionInit(accountId)
	return &AccountSession{SharedSession: *session, provider: sp}, err
}

func (sp *AccountSessionProvider) SessionRead(sessionId string) (*AccountSession, error) {
	session, err := sp.provider.SessionRead(sessionId)
	return &AccountSession{SharedSession: *session, provider: sp}, err
}

func (sp *AccountSessionProvider) SessionDestroy(session *AccountSession) error {
	return sp.provider.SessionDestroy(&session.SharedSession)
}

func (sp *AccountSessionProvider) SessionDestroyOther(session *AccountSession) error {
	return sp.provider.SharedSessionDestroyOther(&session.SharedSession)
}

func (sp *AccountSessionProvider) AllSessions(session *AccountSession) ([]AccountSession, error) {
	allSessions, err := sp.provider.SessionFetchAll(&session.SharedSession)
	if err != nil {
		return nil, err
	}

	accSession := make([]AccountSession, len(allSessions))
	for i, _ := range allSessions {
		accSession[i].provider = sp
		accSession[i].SharedSession = allSessions[i] 
	}
	return accSession, nil
}

func (sp *AccountSessionProvider) Get(id int, key string) (string, error) {
	return sp.provider.Get(id, key)
}

func (sp *AccountSessionProvider) Set(id int, key, value string) error {
	return sp.provider.Set(id, key, value)
}

type AccountSession struct {
	provider *AccountSessionProvider
	entity.SharedSession
}