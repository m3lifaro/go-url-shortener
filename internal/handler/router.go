package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/m3lifaro/go-url-shortener/internal/auth"
	"go.uber.org/zap"
)

func NewRouter(h *Handlers, logger *zap.Logger, auth *auth.AuthImpl) chi.Router {
	r := chi.NewRouter()
	r.Use(gzipMiddleware(logger))
	r.Use(LoggingMiddleware(logger))

	r.Group(func(r chi.Router) {
		r.Get("/ping", h.Ping)
	})

	r.Group(func(r chi.Router) {
		r.Use(authMiddlewareOptional(logger, auth))
		r.Route("/", func(r chi.Router) {
			r.Post("/", h.Shorten)
			r.Post("/api/shorten", h.ShortenJSON)
			r.Post("/api/shorten/batch", h.BatchShorten)
			r.Get("/{id}", h.Redirect)
		})
	})

	// Routes with required auth
	r.Group(func(r chi.Router) {
		r.Use(authMiddlewareOptional(logger, auth))

		r.Route("/api/user", func(r chi.Router) {
			r.Get("/urls", h.User)
			r.Delete("/urls", h.Delete)
		})
	})

	return r
}
