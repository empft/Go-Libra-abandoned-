package main

import (
	"net/http"
	"time"
	"log"
	"context"
	
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/hostrouter"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	r := chi.NewRouter()
	hr := hostrouter.New()

	// A good base middleware stack
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		
		// Set a timeout value on the request context (ctx), that will signal
  		// through ctx.Done() that the request has timed out and further
		// processing should be stopped.
		middleware.Timeout(60 * time.Second),
	)

	hr.Map("localhost:1337", defaultRouter())
  	hr.Map("api.localhost:1337", apiRouter())

	r.Mount("/", hr)

	log.Fatal(http.ListenAndServe(":1337", r))
}

func requestVerifyCode(w http.ResponseWriter, r *http.Request) {

}

func checkUsername(w http.ResponseWriter, r *http.Request) {

}

func createAccount(w http.ResponseWriter, r *http.Request) {

}

func sessionToken(w http.ResponseWriter, r *http.Request) {

}

func apiRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the API"))
	})

	r.Route("/register", func(r chi.Router) {
		r.Post("/", createAccount)
		r.Post("/code", requestVerifyCode)
		r.Post("/username", checkUsername)
	})

	r.Post("login", func )

	return r
}

func defaultRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the main page."))
	})

	return r
}