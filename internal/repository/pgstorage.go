package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m3lifaro/go-url-shortener/internal/model"
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

func (s *PGStorage) Get(key, userID string) (original string, existed bool, isDeleted bool, error error) {
	ctx := context.TODO()
	println("11111")
	var value string
	err := s.pool.QueryRow(ctx, "select original_url, is_deleted from shorten_links where short_url=$1", key).Scan(&value, &isDeleted)
	//err := s.pool.QueryRow(ctx, "select original_url from shorten_links where short_url=$1 and user_id=$2", key, userID).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, isDeleted, nil
		} else {
			s.logger.Error("Failed to get shorten link", zap.String("key", key), zap.Error(err))
			return "", false, isDeleted, err
		}
	}
	return value, true, isDeleted, nil
}
func (s *PGStorage) GetAll(userID string) ([]model.UserLinkDto, error) {
	ctx := context.TODO()

	rows, err := s.pool.Query(ctx, `
        SELECT 
            short_url, 
            original_url
        FROM shorten_links 
        WHERE user_id = $1
    `, userID)

	if err != nil {
		s.logger.Error("Failed to get user links",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var links []model.UserLinkDto
	for rows.Next() {
		var link model.UserLinkDto
		err := rows.Scan(
			&link.ShortURL,
			&link.OriginalURL,
		)
		if err != nil {
			s.logger.Error("Failed to scan link row", zap.Error(err))
			continue
		}
		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error during rows iteration", zap.Error(err))
		return nil, err
	}

	return links, nil
}

func (s *PGStorage) Set(key, url, userID string) (string, error) {
	ctx := context.TODO()
	var existedURL string
	var isNew bool
	err := s.pool.QueryRow(ctx, `
	   INSERT INTO shorten_links(short_url, original_url, user_id)
	   VALUES ($1, $2, $3)
	   ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url -- фиктивное обновление
	   RETURNING short_url, (xmax = 0) AS is_new
	`, key, url, userID).Scan(&existedURL, &isNew)

	if err != nil {
		s.logger.Error("Failed to set shorten link", zap.String("key", key), zap.Error(err))
		return "", err
	}
	if !isNew {
		return existedURL, nil
	}
	return "", nil
}

func (s *PGStorage) BatchSet(records map[string]string, userID string) error {
	ctx := context.TODO()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for key, url := range records {
		if _, err := tx.Exec(ctx,
			"INSERT INTO shorten_links(short_url, original_url, user_id) VALUES ($1, $2, $3)",
			key, url, userID,
		); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *PGStorage) BatchDelete(records []string, userID string) error {
	ctx := context.TODO()

	// Создаем SQL запрос с правильным синтаксисом для массива
	query := `
        UPDATE shorten_links 
        SET is_deleted = true 
        WHERE user_id = $1 AND short_url = ANY($2)`

	// Выполняем запрос
	_, err := s.pool.Exec(ctx, query, userID, records)
	if err != nil {
		return fmt.Errorf("failed to batch delete records: %w", err)
	}

	return nil
}

func (s *PGStorage) Close() error {
	s.pool.Close()
	return nil
}
