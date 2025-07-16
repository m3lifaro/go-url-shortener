package service

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/m3lifaro/go-url-shortener/internal/repository"
)

const defaultLength = 8

type Shortener struct {
	storage repository.Storage
}

func NewShortener(storage repository.Storage) *Shortener {
	return &Shortener{storage: storage}
}

func (s *Shortener) Shorten(url string) (string, error) {
	shortenURL, err := generateRandomString(defaultLength)
	if err != nil {
		return "", err
	}
	s.storage.Set(shortenURL, url)
	return shortenURL, nil
}

func (s *Shortener) GetOriginal(key string) (string, bool) {
	return s.storage.Get(key)
}

func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:n], nil
}
