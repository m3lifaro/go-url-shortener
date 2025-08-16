package main

import (
	"log"
	"net/http"

	"github.com/m3lifaro/go-url-shortener/cmd/config"
	"github.com/m3lifaro/go-url-shortener/internal/handler"
	"github.com/m3lifaro/go-url-shortener/internal/logger"
	"github.com/m3lifaro/go-url-shortener/internal/repository"
	"github.com/m3lifaro/go-url-shortener/internal/service"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Config: %v", cfg)

	zl, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	storage, err := repository.NewMemoryStorage(cfg.StorageFile, zl)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	shortenService := service.NewShortener(storage)
	handlers := handler.NewHandlers(shortenService, cfg.BaseURL, cfg.DBDsn, zl)
	r := handler.NewRouter(handlers, zl)
	log.Printf("Server started on %s", cfg.ServeAddress)
	log.Fatal(http.ListenAndServe(cfg.ServeAddress, r))
}
