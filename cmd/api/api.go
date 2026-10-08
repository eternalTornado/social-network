package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/truongndv/social/internal/store"
)

type application struct {
	config config
	store  store.Storage
}

type config struct {
	addr string
	db   dbConfig
	env  string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)                 // gán cho request 1 id duy nhát
	r.Use(middleware.ClientIPFromRemoteAddr)    // lấy IP của client từ r.RemoteAddr
	r.Use(middleware.Logger)                    // Log in ra cho mỗi request
	r.Use(middleware.Recoverer)                 //bắt panic xảy ra trong handle. In stack trace và return 500
	r.Use(middleware.Timeout(60 * time.Second)) // Đặt livetime 60s vào r.Context(). Khi handle return thì client nhận 504: Gateway Timeout

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func(r chi.Router) {
				// r.Use(app.postsContextMiddleware)

				// r.Get("/", app.getPostHandler)
				// r.Patch("/", app.updatePostHandler)
				// r.Delete("/", app.deletePostHandler)
			})
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	server := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Server has started at %s", app.config.addr)
	return server.ListenAndServe()
}
