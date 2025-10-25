package handler

import (
	"net/http"

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

func NewHandlers(svc *service.Shortener, baseURL string, dsn string, logger *zap.Logger) *Handlers {
	return &Handlers{
		Shorten:      NewShortenHandler(svc, baseURL, logger).ServeHTTP,
		Redirect:     NewRedirectHandler(svc, logger).ServeHTTP,
		ShortenJSON:  NewShortenJSONHandler(svc, baseURL, logger).ServeHTTP,
		User:         NewShortenJSONHandler(svc, baseURL, logger).ServeUserHTTP,
		Delete:       NewShortenJSONHandler(svc, baseURL, logger).ServeDeleteHTTP,
		BatchShorten: NewBatchShortenHandler(svc, baseURL, logger).ServeHTTP,
		Ping:         NewPingHandler(dsn, logger).ServeHTTP,
	}
}
