package handler

import (
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"io"
	"net/http"
)

type ShortenHandler struct {
	service *service.Shortener
}

func NewShortenHandler(service *service.Shortener) *ShortenHandler {
	return &ShortenHandler{service: service}
}

func (h *ShortenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	url := string(body)
	shortedURL := h.service.Shorten(url)
	println("URL: " + url)
	println("Shorten url: " + shortedURL)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://localhost:8080/" + shortedURL))
}
