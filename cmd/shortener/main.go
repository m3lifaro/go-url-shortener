package main

import (
	"github.com/m3lifaro/go-url-shortener/cmd/config"
	"github.com/m3lifaro/go-url-shortener/internal/handler"
	"github.com/m3lifaro/go-url-shortener/internal/repository"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Println("Config:", cfg)
	storage := repository.NewMemoryStorage()
	shortenService := service.NewShortener(storage)
	handlers := handler.NewHandlers(shortenService, cfg.BaseURL)
	r := handler.NewRouter(handlers)
	log.Printf("Server started on %s", cfg.ServeAddress)
	log.Fatal(http.ListenAndServe(cfg.ServeAddress, r))
}
