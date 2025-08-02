package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/m3lifaro/go-url-shortener/internal/logger"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type RedirectHandler struct {
	service *service.Shortener
}

func NewRedirectHandler(service *service.Shortener) *RedirectHandler {
	return &RedirectHandler{service: service}
}

func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("[Redirect handler] Handle event")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	key := chi.URLParam(r, "id")
	url, exists, err := h.service.GetOriginal(key)
	if err != nil {
		logger.Log.Error("got error getting original",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	log.Println("Redirecting to: " + url)

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
