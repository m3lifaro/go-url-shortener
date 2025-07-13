package handler

import (
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"net/http"
)

type RedirectHandler struct {
	service *service.Shortener
}

func NewRedirectHandler(service *service.Shortener) *RedirectHandler {
	return &RedirectHandler{service: service}
}

func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	key := r.PathValue("id")
	url, exists := h.service.GetOriginal(key)
	if !exists {
		//w.WriteHeader(http.StatusNotFound)
		http.NotFound(w, r)
		return
	}
	println("Redirecting to: " + url)

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
