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

func (s *PGStorage) Set(key, url string) (string, error) {
	ctx := context.TODO()
	var existedURL string
	var isNew bool
	err := s.pool.QueryRow(ctx, `
	   INSERT INTO shorten_links(short_url, original_url)
	   VALUES ($1, $2)
	   ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url -- фиктивное обновление
	   RETURNING short_url, (xmax = 0) AS is_new
	`, key, url).Scan(&existedURL, &isNew)
	//	err := s.pool.QueryRow(ctx, `
	//    WITH insert_attempt AS (
	//        INSERT INTO shorten_links (short_url, original_url)
	//        VALUES ($1, $2)
	//        ON CONFLICT (original_url) DO NOTHING
	//        RETURNING short_url, true AS is_new
	//    )
	//    SELECT short_url, is_new FROM insert_attempt
	//    UNION ALL
	//    SELECT short_url, false FROM shorten_links WHERE original_url = $2 AND NOT EXISTS (SELECT 1 FROM insert_attempt)
	//    LIMIT 1
	//`, key, url).Scan(&existedUrl, &isNew)
	//_, err := s.pool.Exec(ctx, "insert into shorten_links(short_url, original_url) values ($1, $2)", key, url)
	if err != nil {
		s.logger.Error("Failed to set shorten link", zap.String("key", key), zap.Error(err))
		return "", err
	}
	if !isNew {
		return existedURL, nil
	}
	return "", nil
}

func (s *PGStorage) BatchSet(records map[string]string) error {
	ctx := context.TODO()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for key, url := range records {
		if _, err := tx.Exec(ctx,
			"INSERT INTO shorten_links(short_url, original_url) VALUES ($1, $2)",
			key, url,
		); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *PGStorage) Close() error {
	s.pool.Close()
	return nil
}
