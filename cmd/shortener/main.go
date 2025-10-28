package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/m3lifaro/go-url-shortener/cmd/config"
	"github.com/m3lifaro/go-url-shortener/internal/audit"
	shortenAuth "github.com/m3lifaro/go-url-shortener/internal/auth"
	"github.com/m3lifaro/go-url-shortener/internal/handler"
	"github.com/m3lifaro/go-url-shortener/internal/logger"
	"github.com/m3lifaro/go-url-shortener/internal/repository"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"github.com/pressly/goose/v3"
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

	var storage repository.Storage

	if cfg.DBDsn != "" {
		pool, err := pgxpool.New(context.Background(), cfg.DBDsn)
		if err != nil {
			log.Fatalf("Failed to initialize connection pool: %v", err)
		}

		db := stdlib.OpenDBFromPool(pool)

		if err := goose.Up(db, "migrations"); err != nil {
			log.Fatal("goose up failed:", err)
		}

		log.Println("Migrations applied successfully")

		if err := db.Close(); err != nil {
			log.Fatal("DB wasn't closed:", err)
		}
		storage = repository.NewPGStorage(pool, zl)
	} else {
		storage, err = repository.NewMemoryStorage(cfg.StorageFile, zl)
		if err != nil {
			log.Fatalf("Failed to initialize storage: %v", err)
		}
	}
	defer storage.Close()

	auditMgr := &audit.Manager{}
	if cfg.AuditFile != "" {
		auditMgr.Register(audit.NewFileSubscriber(cfg.AuditFile))
	}
	if cfg.AuditURL != "" {
		auditMgr.Register(audit.NewHTTPSubscriber(cfg.AuditURL))
	}

	shortenService := service.NewShortener(storage)
	handlers := handler.NewHandlers(shortenService, cfg.BaseURL, cfg.DBDsn, zl, auditMgr)
	auth := shortenAuth.NewAuth(cfg.AuthSecret)
	r := handler.NewRouter(handlers, zl, auth)
	log.Printf("Server started on %s", cfg.ServeAddress)
	log.Fatal(http.ListenAndServe(cfg.ServeAddress, r))
}
