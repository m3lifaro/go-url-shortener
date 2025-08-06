package handler

import (
	"encoding/json"
	"fmt"
	"github.com/m3lifaro/go-url-shortener/internal/model"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"go.uber.org/zap"
	"mime"
	"net/http"
)

const jsonContentType = "application/json"

type ShortenJSONHandler struct {
	service *service.Shortener
	baseURL string
	logger  *zap.Logger
}

func NewShortenJSONHandler(service *service.Shortener, baseURL string, logger *zap.Logger) *ShortenJSONHandler {
	return &ShortenJSONHandler{service: service, baseURL: baseURL, logger: logger}
}

func (h *ShortenJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req model.ShortenRequest

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&req); err != nil {
		h.logger.Error(
			"got error, while decoding HTTP request",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	contentHeader := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentHeader)
	if err != nil || mediaType != jsonContentType {
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
		w.WriteHeader(http.StatusInternalServerError)
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

	resp := model.ShortenResponse{Result: respURL}
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		h.logger.Debug("error encoding response", zap.Error(err))
		return
	}
}
