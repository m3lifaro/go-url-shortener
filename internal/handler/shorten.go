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
	println("shorten handler")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//contentHeader := r.Header.Get("content-type")
	//if contentHeader != "text/plain" {
	//	println("content-type is not text/plain. [func (h *ShortenHandler) ServeHTTP]")
	//	w.WriteHeader(http.StatusBadRequest)
	//	w.Write([]byte("Unsupported content-type. Only 'text/plain' allowed"))
	//	return
	//}
	defer r.Body.Close()

	url := string(body)
	if len(url) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Empty url not allowed"))
		return
	}
	shortedURL := h.service.Shorten(url)
	println("URL: " + url)
	println("Shorten url: " + shortedURL)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://localhost:8080/" + shortedURL))
}
