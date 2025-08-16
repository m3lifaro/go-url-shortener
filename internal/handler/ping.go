package handler

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type PingHandler struct {
	dsn    string
	logger *zap.Logger
}

func NewPingHandler(dsn string, logger *zap.Logger) *PingHandler {
	return &PingHandler{dsn: dsn, logger: logger}
}

func (h *PingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, h.dsn)
	if err != nil {
		h.logger.Error(
			"got error while connecting to database",
			zap.Error(err),
		)
		http.Error(w, "Database unavailable", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	if err := conn.Ping(ctx); err != nil {
		h.logger.Error(
			"got error while ping database",
			zap.Error(err),
		)
		http.Error(w, "Database unavailable", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}
