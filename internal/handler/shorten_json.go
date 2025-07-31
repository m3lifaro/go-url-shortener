package handler

import (
	"encoding/json"
	"fmt"
	"github.com/m3lifaro/go-url-shortener/internal/logger"
	"github.com/m3lifaro/go-url-shortener/internal/model"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"go.uber.org/zap"
	"log"
	"mime"
	"net/http"
	"os"
)

const jsonContentType = "application/json"

type ShortenJSONHandler struct {
	service *service.Shortener
	baseURL string
}

func NewShortenJSONHandler(service *service.Shortener, baseURL string) *ShortenJSONHandler {
	return &ShortenJSONHandler{service: service, baseURL: baseURL}
}

func (h *ShortenJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("[Shorten JSON handler] Handle event")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req model.ShortenRequest

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&req); err != nil {
		logger.Log.Error("got error, while decoding HTTP request",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	contentHeader := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentHeader)
	if err != nil || mediaType != jsonContentType {
		log.Println("Content-Type is not application/json. [func (h *ShortenHandler) ServeHTTP]")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported Content-Type. Expected 'application/json', got: " + mediaType))
		return
	}
	defer r.Body.Close()

	url := req.ShortingURL
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

	resp := model.ShortenResponse{Result: fmt.Sprintf("%s%s", h.baseURL, shortedURL)}
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		return
	}
}
