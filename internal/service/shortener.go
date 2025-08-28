package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/m3lifaro/go-url-shortener/internal/model"
	"github.com/m3lifaro/go-url-shortener/internal/repository"
)

const defaultLength = 8

type Shortener struct {
	storage repository.Storage
}

func NewShortener(storage repository.Storage) *Shortener {
	return &Shortener{storage: storage}
}

func (s *Shortener) Shorten(url string, userID string) (shorten string, existed bool, err error) {
	shortenURL, err := generateRandomString(defaultLength)
	if err != nil {
		return "", false, err
	}
	existedURL, err := s.storage.Set(shortenURL, url, userID)
	if err != nil {
		return "", false, err
	}
	if existedURL != "" {
		return existedURL, true, nil
	}
	return shortenURL, false, nil
}

func (s *Shortener) BatchShorten(urls []model.BatchRequestItem, baseURL, userID string) ([]model.BatchResponseItem, error) {
	records := make(map[string]string)
	response := make([]model.BatchResponseItem, 0, len(urls))

	for _, item := range urls {
		shortURL, err := generateRandomString(defaultLength)
		if err != nil {
			return nil, err
		}
		records[shortURL] = item.OriginalURL
		response = append(response, model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      fmt.Sprintf("%s%s", baseURL, shortURL),
		})
	}

	if err := s.storage.BatchSet(records, userID); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Shortener) GetOriginal(key, userID string) (string, bool, error) {
	return s.storage.Get(key, userID)
}

func (s *Shortener) GetUserUrls(userID string) ([]model.UserResponseItem, error) {
	return s.storage.GetAll(userID)
}

func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:n], nil
}
