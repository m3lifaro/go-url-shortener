package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/m3lifaro/go-url-shortener/internal/service"
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
	url, exists := h.service.GetOriginal(key)
	if !exists {
		http.NotFound(w, r)
		return
	}
	log.Println("Redirecting to: " + url)

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
