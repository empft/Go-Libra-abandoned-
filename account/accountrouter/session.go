package accountrouter

import (
	"context"
	"errors"
	"net/http"

	"github.com/stevealexrs/Go-Libra/namespace/cookiens"
	"github.com/stevealexrs/Go-Libra/session"
)

type userSession struct {
	session.Shared
	provider session.SharedProvider
}

type businessSession struct {
	session.Shared
	provider session.SharedProvider
}

func (s *userSession) Attach(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: cookiens.UserSession,
		Value: s.SessionId(),
	})
}

func (rt *Router) readUserSession(ctx context.Context, r *http.Request) (*userSession, error) {
	cookie, err := r.Cookie(cookiens.UserSession)
	if errors.Is(err, http.ErrNoCookie) {
		return nil, errors.New("user session cookie is not set")
	} else if err != nil {
		return nil, err
	}

	shared, err := rt.userProvider.Read(ctx, cookie.Value)
	if err != nil {
		return nil, err
	}

	return rt.userSession(shared), nil
}

func (s *businessSession) Attach(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: cookiens.BusinessSession,
		Value: s.SessionId(),
	})
}

