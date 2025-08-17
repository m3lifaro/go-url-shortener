package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PGStorage struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewPGStorage(pool *pgxpool.Pool, logger *zap.Logger) *PGStorage {
	return &PGStorage{
		pool:   pool,
		logger: logger,
	}
}

func (s *PGStorage) Get(key string) (string, bool, error) {
	ctx := context.TODO()
	var value string
	err := s.pool.QueryRow(ctx, "select original_url from shorten_links where short_url=$1", key).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		} else {
			s.logger.Error("Failed to get shorten link", zap.String("key", key), zap.Error(err))
			return "", false, err
		}
	}
	return value, true, nil
}

func (s *PGStorage) Set(key, url string) error {
	ctx := context.TODO()
	_, err := s.pool.Exec(ctx, "insert into shorten_links(short_url, original_url) values ($1, $2)", key, url)
	if err != nil {
		s.logger.Error("Failed to set shorten link", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

func (s *PGStorage) Close() error {
	s.pool.Close()
	return nil
}
