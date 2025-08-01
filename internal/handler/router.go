package handler

import (
	"github.com/go-chi/chi/v5"
	"log"
)

func NewRouter(h *Handlers) chi.Router {
	r := chi.NewRouter()
	r.Use(gzipMiddleware)
	r.Use(LoggingMiddleware)
	r.Route("/", func(r chi.Router) {
		r.Post("/", h.Shorten)
		r.Post("/api/shorten", h.ShortenJSON)
		r.Get("/{id}", h.Redirect)
	})
	log.Println("Hello from CHI Router")
	return r
}
