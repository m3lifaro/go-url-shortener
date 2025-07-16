package handler

import (
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"net/http"
)

type Handlers struct {
	Shorten  http.HandlerFunc
	Redirect http.HandlerFunc
}

func NewHandlers(svc *service.Shortener, baseURL string) *Handlers {
	return &Handlers{
		Shorten:  NewShortenHandler(svc, baseURL).ServeHTTP,
		Redirect: NewRedirectHandler(svc).ServeHTTP,
	}
}
