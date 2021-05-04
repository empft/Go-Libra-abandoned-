package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/hostrouter"
)


func masterRouter() chi.Router {
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
		middleware.Timeout(60*time.Second),
	)

	hr.Map("localhost:1337", defaultRouter())
	hr.Map("api.localhost:1337", apiRouter())

	r.Mount("/", hr)
	return r
}

func apiRouter() chi.Router {
	r := chi.NewRouter()

	return r
}

func defaultRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the main page."))
	})

	return r
}