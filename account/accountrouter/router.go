package accountrouter

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/stevealexrs/Go-Libra/account"
	"github.com/stevealexrs/Go-Libra/mware"
	"github.com/stevealexrs/Go-Libra/session"
)

type Router struct {
	user 	 		 account.UserCreator
	userProvider 	 session.SharedProvider
	userRecovery	 account.UserAccountRecoveryHelper
	business 		 account.BusinessCreator
	businessProvider session.SharedProvider
	businessRecovery account.BusinessAccountRecoveryHelper
}

func New(
	user account.UserCreator,
	userProvider session.SharedProvider,
	userRecovery account.UserAccountRecoveryHelper,
	business account.BusinessCreator,
	businessProvider session.SharedProvider,
	businessRecovery account.BusinessAccountRecoveryHelper,
	) *Router {
	return &Router{
		user: user,
		userProvider: userProvider,
		userRecovery: userRecovery,
		business: business,
		businessProvider: businessProvider,
		businessRecovery: businessRecovery,
	}
}

func (rt *Router) useDefaultMiddlewares(r *chi.Mux) {
	r.Use(
		mware.Localization,
		// Minimum global rate limit
		httprate.Limit(
			10,
			10*time.Second,
			httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
		),
	)
}

func (rt *Router) UserHandler() chi.Router {
	r := chi.NewRouter()
	
	rt.useDefaultMiddlewares(r)

	// Ip only rate limit
	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, 30*time.Minute))

		r.Get("/exist", errorHandler(rt.userExists()))
		r.Post("/register", errorHandler(rt.userExists()))
		r.Post("/register-with-invitation", errorHandler(rt.userRegisterWithInvitation()))
		r.Post("/reset-password", errorHandler(rt.userResetPassword()))
	})

	// Email rate limit
	r.Group(func(r chi.Router) {
		r.Use(httprate.Limit(
			5,
			time.Hour,
			httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
				err := r.ParseForm()
				if err != nil {
					return "", err
				}

				return r.PostForm.Get("email"), nil
			}),
		))

		r.Post("/invitation", errorHandler(rt.userInvitation()))
		r.Post("/forget-username", errorHandler(rt.userForgetUsername()))
		r.Post("/forget-password", errorHandler(rt.userForgetPassword()))
	})

	// login rate limit
	r.Group(func(r chi.Router) {
		r.Use(httprate.Limit(
			10,
			30*time.Minute,
			httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
				err := r.ParseForm()
				if err != nil {
					return "", err
				}

				return r.PostForm.Get("username"), nil
			}, httprate.KeyByIP),
		))

		r.Post("/login", errorHandler(rt.userLogin()))
	})
	
	r.Get("/logout", errorHandler(rt.userLogout()))

}

func (rt *Router) BusinessHandler() chi.Router {
	r := chi.NewRouter()

	rt.useDefaultMiddlewares(r)

	// Ip only rate limit
	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, 30*time.Minute))

		r.Get("/exist", errorHandler(rt.businessExists()))
		r.Post("/register", errorHandler(rt.businessRegister()))
		r.Post("/register-with-identity", errorHandler(rt.businessRegisterWithIdentity()))
		r.Post("/reset-password", errorHandler(rt.businessResetPassword()))
	})

	// Email rate limit
	r.Group(func(r chi.Router) {
		r.Use(httprate.Limit(
			5,
			time.Hour,
			httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
				err := r.ParseForm()
				if err != nil {
					return "", err
				}

				return r.PostForm.Get("email"), nil
			}),
		))

		r.Post("/forget-username", errorHandler(rt.businessForgetUsername()))
		r.Post("/forget-password", errorHandler(rt.businessForgetPassword()))
	})

	// login rate limit
	r.Group(func(r chi.Router) {
		r.Use(httprate.Limit(
			10,
			30*time.Minute,
			httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
				err := r.ParseForm()
				if err != nil {
					return "", err
				}

				return r.PostForm.Get("username"), nil
			}, httprate.KeyByIP),
		))

		r.Post("/login", errorHandler(rt.businessLogin()))
	})

	r.Get("/logout", errorHandler(rt.businessLogout()))
	return r
}