package handler

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func NewRouter(h *Handlers, logger *zap.Logger) chi.Router {
	r := chi.NewRouter()
	r.Use(gzipMiddleware(logger))
	r.Use(LoggingMiddleware(logger))
	r.Route("/", func(r chi.Router) {
		r.Post("/", h.Shorten)
		r.Post("/api/shorten", h.ShortenJSON)
		r.Get("/{id}", h.Redirect)
	})
	return r
}
