package handler

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/m3lifaro/go-url-shortener/internal/audit"
	"github.com/m3lifaro/go-url-shortener/internal/auth"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"go.uber.org/zap"
)

const contentType = "text/plain"

type ShortenHandler struct {
	service      *service.Shortener
	baseURL      string
	logger       *zap.Logger
	auditManager *audit.Manager
}

func NewShortenHandler(service *service.Shortener, baseURL string, logger *zap.Logger, auditManager *audit.Manager) *ShortenHandler {
	return &ShortenHandler{service: service, baseURL: baseURL, logger: logger, auditManager: auditManager}
}

func (h *ShortenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), defaultTimeoutSec*time.Second)
	defer cancel()
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

	contentHeader := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentHeader)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid Content-Type header"))
		return
	}

	if mediaType != "text/plain" && mediaType != "application/x-gzip" && mediaType != "plain/text" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported Content-Type. Expected 'text/plain', 'plain/text' or 'application/x-gzip', got: " + mediaType))
		return
	}

	url := string(body)
	if len(url) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Empty url not allowed"))
		return
	}
	userID, _ := auth.GetUserID(r.Context())

	event := audit.Event{
		Timestamp: time.Now().Unix(),
		Action:    "shorten",
		UserID:    userID,
		URL:       url,
	}
	h.auditManager.NotifyAll(event)

	h.logger.Info("Shorten cookie context", zap.String("user_id", userID))
	shortedURL, existedURL, err := h.service.Shorten(ctx, url, userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.logger.Error(
			"got error while shortening url",
			zap.Error(err),
		)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	w.Header().Set("Content-Type", contentType)
	if existedURL {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	var respURL = fmt.Sprintf("%s%s", h.baseURL, shortedURL)
	h.logger.Debug("Shorten params",
		zap.String("url", url),
		zap.String("shortedURL", respURL),
		zap.Bool("existed", existedURL),
		zap.String("user_id", userID),
	)

	w.Write([]byte(respURL))
}
