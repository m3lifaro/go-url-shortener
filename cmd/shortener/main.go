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
	handlers := handler.NewHandlers(shortenService)
	r := handler.NewRouter(handlers)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
