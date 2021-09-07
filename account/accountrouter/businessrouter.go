package accountrouter

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/h2non/filetype"
	"github.com/stevealexrs/Go-Libra/account"
	"github.com/stevealexrs/Go-Libra/namespace/cookiens"
	"github.com/stevealexrs/Go-Libra/session"
)

func (rt *Router) businessSession(shared *session.Shared) *businessSession {
	return &businessSession{
		*shared,
		rt.businessProvider,
	}
}

func (rt *Router) businessExists() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		name := r.URL.Query().Get("username")

		exist, err := rt.business.UsernameExist(r.Context(), name)
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

func (rt *Router) businessRegister() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}

		id, err := rt.business.CreateAccount(r.Context(), account.BusinessRegistrationForm{
			Username:    r.PostForm.Get("username"),
			DisplayName: r.PostForm.Get("displayName"),
			Password:    r.PostForm.Get("password"),
			Email:       r.PostForm.Get("email"),
		})
		if err != nil {
			return err
		}

		shared, err := rt.businessProvider.Init(r.Context(), id)
		if err != nil {
			return err
		}

		rt.businessSession(shared).Attach(w)
		return nil
	}
}

type AccountForm struct {

}

func (rt *Router) businessRegisterWithIdentity() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		r.Body = http.MaxBytesReader(w, r.Body, int64(math.Pow10(7)))
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}

		files := make([][]byte, 0)
		fhs := r.MultipartForm.File["document"]
		if len(fhs) > account.MaxBusinessDocuments {
			return account.ErrTooManyFiles(r.Context())
		}
		for _, v := range fhs {
			if v.Size > account.MaxBusinessDocumentSize {
				return account.ErrFileTooLarge(r.Context())
			}

			f, err := v.Open()
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			_, err = io.Copy(&buf, f)
			if err != nil {
				return err
			}

			if !(filetype.IsImage(buf.Bytes()) || filetype.Is(buf.Bytes(), "pdf")) {
				return account.ErrInvalidFileType(r.Context())
			}

			files = append(files, buf.Bytes())
		}

		id, err := rt.business.CreateAccountWithIdentity(r.Context(), account.BusinessRegistrationFormWithIdentity{
			BusinessRegistrationForm: account.BusinessRegistrationForm{
				Username:    r.PostFormValue("username"),
				DisplayName: r.PostFormValue("displayName"),
				Password:    r.PostFormValue("password"),
				Email:       r.PostFormValue("email"),
			},
			OfficialName: 		r.PostFormValue("businessName"),
			RegistrationNumber: r.PostFormValue("registrationNumber"),
			Address: 			r.PostFormValue("address"),
			Documents: 			files,
		})
		if err != nil {
			return err
		}

		shared, err := rt.businessProvider.Init(r.Context(), id)
		if err != nil {
			return err
		}

		rt.businessSession(shared).Attach(w)
		return nil
	}
}

func (rt *Router) businessLogin() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}

		id, err := rt.business.Login(r.Context(), r.PostForm.Get("username"), r.PostForm.Get("password"))
		if err != nil {
			return err
		}

		shared, err := rt.businessProvider.Init(r.Context(), id)
		if err != nil {
			return err
		}
		rt.businessSession(shared).Attach(w)
		return nil
	}
}

func (rt *Router) businessForgetUsername() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}

		return rt.businessRecovery.RequestUsernameReminder(r.Context(), r.PostForm.Get("email"))
	}
}

func (rt *Router) businessForgetPassword() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}

		return rt.businessRecovery.RequestPasswordReset(r.Context(), r.PostForm.Get("username"), r.PostForm.Get("email"))
	}
}

func (rt *Router) businessResetPassword() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}

		return rt.businessRecovery.ResetPassword(r.Context(), r.PostForm.Get("token"), r.PostForm.Get("password"))
	}
}

func (rt *Router) businessLogout() errorFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		cookie, err := r.Cookie(cookiens.BusinessSession)
		if err != nil {
			return err
		}

		err = rt.businessProvider.Destroy(r.Context(), cookie.Value)
		if err != nil {
			return err
		}

		cookie.Expires = time.Unix(0, 0)
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)

		return nil
	}
}