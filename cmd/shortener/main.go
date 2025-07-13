package main

import (
	"github.com/m3lifaro/go-url-shortener/internal/handler"
	"github.com/m3lifaro/go-url-shortener/internal/repository"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"log"
	"net/http"
)

func main() {
	storage := repository.NewMemoryStorage()
	shortenService := service.NewShortener(storage)
	redirectHandler := handler.NewRedirectHandler(shortenService)
	shortenHandler := handler.NewShortenHandler(shortenService)

	mux := http.NewServeMux()
	mux.Handle("GET /{id}", redirectHandler)
	mux.Handle("POST /", shortenHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
