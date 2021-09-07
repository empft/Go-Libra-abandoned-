package accountrouter

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/stevealexrs/Go-Libra/account"
	"github.com/stevealexrs/Go-Libra/namespace/cookiens"
	"github.com/stevealexrs/Go-Libra/session"
)

func (rt *Router) userSession(shared *session.Shared) *userSession {
	return &userSession{
		*shared,
		rt.userProvider,
	}
}

func (rt *Router) userExists() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		name := r.URL.Query().Get("username")

		exist, err := rt.user.UsernameExist(r.Context(), name)
		if err != nil {
			return err
		}

		res, err := json.Marshal(exist)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
		return nil
	}
}

func (rt *Router) userRegister() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}

		id, err := rt.user.CreateAccount(r.Context(), account.UserRegistrationForm{
			Invitation:  account.InvitationEmail{},
			Username:    r.PostForm.Get("username"),
			DisplayName: r.PostForm.Get("displayName"),
			Password:    r.PostForm.Get("password"),
			Email:       r.PostForm.Get("email"),
		})
		if err != nil {
			return err
		}

		shared, err := rt.userProvider.Init(r.Context(), id)
		if err != nil {
			return err
		}

		rt.userSession(shared).Attach(w)
		return nil
	}
}

func (rt *Router) userInvitation() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}

		return rt.user.CreateInvitation(r.Context(), r.PostForm.Get("email"))
	}
}

func (rt *Router) userRegisterWithInvitation() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}

		id, err := rt.user.CreateAccountWithInvitation(r.Context(), account.UserRegistrationForm{
			Invitation: account.InvitationEmail{
				Email: r.PostForm.Get("invitationEmail"),
				Code:  r.PostForm.Get("invitationCode"),
			},
			Username:    r.PostForm.Get("username"),
			DisplayName: r.PostForm.Get("displayName"),
			Password:    r.PostForm.Get("password"),
			Email:       r.PostForm.Get("email"),
		})
		if err != nil {
			return err
		}

		shared, err := rt.userProvider.Init(r.Context(), id)
		if err != nil {
			return err
		}

		rt.userSession(shared).Attach(w)
		return nil
	}
}

func (rt *Router) userLogin() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}

		id, err := rt.user.Login(r.Context(), r.PostForm.Get("username"), r.PostForm.Get("password"))
		if err != nil {
			return err
		}

		shared, err := rt.userProvider.Init(r.Context(), id)
		if err != nil {
			return err
		}
		rt.userSession(shared).Attach(w)
		return nil
	}
}

func (rt *Router) userForgetUsername() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}

		return rt.userRecovery.RequestUsernameReminder(r.Context(), r.PostForm.Get("email"))
	}
}

func (rt *Router) userForgetPassword() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}

		return rt.userRecovery.RequestPasswordReset(r.Context(), r.PostForm.Get("username"), r.PostForm.Get("email"))
	}
}

func (rt *Router) userResetPassword() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}

		return rt.userRecovery.ResetPassword(r.Context(), r.PostForm.Get("token"), r.PostForm.Get("password"))
	}
}

func (rt *Router) userLogout() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		cookie, err := r.Cookie(cookiens.UserSession)
		if err != nil {
			return err
		}

		err = rt.userProvider.Destroy(r.Context(), cookie.Value)
		if err != nil {
			return err
		}

		cookie.Expires = time.Unix(0, 0)
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)

		return nil
	}
}