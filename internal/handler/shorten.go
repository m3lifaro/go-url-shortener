package handler

import (
	"fmt"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
)

const contentType = "text/plain"

type ShortenHandler struct {
	service *service.Shortener
	baseURL string
}

func NewShortenHandler(service *service.Shortener, baseURL string) *ShortenHandler {
	return &ShortenHandler{service: service, baseURL: baseURL}
}

func (h *ShortenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("[Shorten handler] Handle event")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	contentHeader := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentHeader)
	if err != nil || mediaType != contentType {
		log.Println("Content-Type is not text/plain. [func (h *ShortenHandler) ServeHTTP]")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported Content-Type. Expected 'text/plain', got: " + mediaType))
		return
	}
	defer r.Body.Close()

	url := string(body)
	if len(url) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Empty url not allowed"))
		return
	}
	shortedURL, err := h.service.Shorten(url)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(os.Stderr, "Got error while shortening url: %v\n", err)
		w.Write([]byte("Got error while shortening url: " + err.Error()))
		return
	}
	log.Println("URL: " + url)
	log.Println("Shorten url: " + h.baseURL + shortedURL)

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s%s", h.baseURL, shortedURL)))
}
