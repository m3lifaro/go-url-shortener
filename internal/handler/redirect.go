package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/m3lifaro/go-url-shortener/internal/audit"
	"github.com/m3lifaro/go-url-shortener/internal/auth"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"go.uber.org/zap"
)

type RedirectHandler struct {
	service      *service.Shortener
	logger       *zap.Logger
	auditManager *audit.Manager
}

func NewRedirectHandler(service *service.Shortener, logger *zap.Logger, auditManager *audit.Manager) *RedirectHandler {
	return &RedirectHandler{service: service, logger: logger, auditManager: auditManager}
}

func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), defaultTimeoutSec*time.Second)
	defer cancel()
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	key := chi.URLParam(r, "id")
	userID, _ := auth.GetUserID(r.Context())

	h.logger.Debug("Shorten redirect request details", zap.String("user_id", userID))
	url, exists, deleted, err := h.service.GetOriginal(ctx, key, userID)
	if err != nil {
		h.logger.Error(
			"got error getting original",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if deleted {
		w.WriteHeader(http.StatusGone)
		return
	}
	h.logger.Debug("Redirect to",
		zap.String("url", url),
	)
	event := audit.Event{
		Timestamp: time.Now().Unix(),
		Action:    "shorten",
		UserID:    userID,
		URL:       url,
	}
	h.auditManager.NotifyAll(event)
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
