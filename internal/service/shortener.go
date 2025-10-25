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

func (s *Shortener) GetOriginal(key, userID string) (original string, existed bool, isDeleted bool, error error) {
	return s.storage.Get(key, userID)
}

func (s *Shortener) GetUserUrls(userID string, baseURL string) ([]model.UserResponseItem, error) {
	dtos, err := s.storage.GetAll(userID)
	if err != nil {
		return nil, fmt.Errorf("error getting shortener urls by user(%s): %w", userID, err)
	}
	response := make([]model.UserResponseItem, 0, len(dtos))
	for _, result := range dtos {
		response = append(response, model.UserResponseItem{
			OriginalURL: result.OriginalURL,
			ShortURL:    fmt.Sprintf("%s%s", baseURL, result.ShortURL),
		})
	}
	return response, nil
}

func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:n], nil
}

func (s *Shortener) DeleteUserUrls(userID string, linksToDelete []string) error {
	err := s.storage.BatchDelete(linksToDelete, userID)
	if err != nil {
		return fmt.Errorf("error delete shorten urls by user(%s): %w", userID, err)
	}
	return nil
}
