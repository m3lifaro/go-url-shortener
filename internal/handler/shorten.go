package handler

import (
	"fmt"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"go.uber.org/zap"
	"io"
	"mime"
	"net/http"
)

const contentType = "text/plain"

type ShortenHandler struct {
	service *service.Shortener
	baseURL string
	logger  *zap.Logger
}

func NewShortenHandler(service *service.Shortener, baseURL string, logger *zap.Logger) *ShortenHandler {
	return &ShortenHandler{service: service, baseURL: baseURL, logger: logger}
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
	contentHeader := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentHeader)
	url := string(body)

	if err != nil || (mediaType != "text/plain" && mediaType != "application/x-gzip") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported Content-Type. Expected 'text/plain' or 'application/x-gzip', got: " + mediaType))
		return
	}
	defer r.Body.Close()

	if len(url) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Empty url not allowed"))
		return
	}
	shortedURL, err := h.service.Shorten(url)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.logger.Error(
			"got error while shortening url",
			zap.Error(err),
		)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	var respURL = fmt.Sprintf("%s%s", h.baseURL, shortedURL)
	h.logger.Debug("Shorten params",
		zap.String("url", url),
		zap.String("shortedURL", respURL),
	)

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(respURL))
}
