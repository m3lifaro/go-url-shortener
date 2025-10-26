package handler

import (
	"net/http"

	"github.com/m3lifaro/go-url-shortener/internal/audit"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"go.uber.org/zap"
)

const defaultTimeoutSec = 30

type Handlers struct {
	Shorten      http.HandlerFunc
	Redirect     http.HandlerFunc
	ShortenJSON  http.HandlerFunc
	User         http.HandlerFunc
	Delete       http.HandlerFunc
	BatchShorten http.HandlerFunc
	Ping         http.HandlerFunc
}

func NewHandlers(svc *service.Shortener, baseURL string, dsn string, logger *zap.Logger, auditManager *audit.Manager) *Handlers {
	return &Handlers{
		Shorten:      NewShortenHandler(svc, baseURL, logger, auditManager).ServeHTTP,
		Redirect:     NewRedirectHandler(svc, logger, auditManager).ServeHTTP,
		ShortenJSON:  NewShortenJSONHandler(svc, baseURL, logger, auditManager).ServeHTTP,
		User:         NewShortenJSONHandler(svc, baseURL, logger, auditManager).ServeUserHTTP,
		Delete:       NewShortenJSONHandler(svc, baseURL, logger, auditManager).ServeDeleteHTTP,
		BatchShorten: NewBatchShortenHandler(svc, baseURL, logger).ServeHTTP,
		Ping:         NewPingHandler(dsn, logger).ServeHTTP,
	}
}
