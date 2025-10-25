package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"time"

	"github.com/m3lifaro/go-url-shortener/internal/auth"
	"github.com/m3lifaro/go-url-shortener/internal/model"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"go.uber.org/zap"
)

type BatchShortenHandler struct {
	service *service.Shortener
	baseURL string
	logger  *zap.Logger
}

func NewBatchShortenHandler(service *service.Shortener, baseURL string, logger *zap.Logger) *BatchShortenHandler {
	return &BatchShortenHandler{service: service, baseURL: baseURL, logger: logger}
}

func (h *BatchShortenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), defaultTimeoutSec*time.Second)
	defer cancel()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	contentHeader := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentHeader)
	if err != nil || mediaType != jsonContentType {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported Content-Type. Expected 'application/json'"))
		return
	}

	var batchReq []model.BatchRequestItem
	if err := json.NewDecoder(r.Body).Decode(&batchReq); err != nil {
		h.logger.Error("Failed to decode batch request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	for _, item := range batchReq {
		if item.OriginalURL == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Empty URL for correlation_id: %s", item.CorrelationID)))
			return
		}
	}
	userID, _ := auth.GetUserID(r.Context())

	results, err := h.service.BatchShorten(ctx, batchReq, h.baseURL, userID)
	if err != nil {
		h.logger.Error("Failed to shorten batch request", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		h.logger.Error("Failed to encode batch response", zap.Error(err))
	}
}
