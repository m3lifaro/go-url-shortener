package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"go.uber.org/zap"
	"net/http"
)

type RedirectHandler struct {
	service *service.Shortener
	logger  *zap.Logger
}

func NewRedirectHandler(service *service.Shortener, logger *zap.Logger) *RedirectHandler {
	return &RedirectHandler{service: service, logger: logger}
}

func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	key := chi.URLParam(r, "id")
	url, exists, err := h.service.GetOriginal(key)
	if err != nil {
		h.logger.Error(
			"got error getting original",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	h.logger.Debug("Redirect to",
		zap.String("url", url),
	)

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
