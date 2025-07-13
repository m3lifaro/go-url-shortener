package service

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/m3lifaro/go-url-shortener/internal/repository"
)

type Shortener struct {
	storage repository.Storage
}

func NewShortener(storage repository.Storage) *Shortener {
	return &Shortener{storage: storage}
}

func (s *Shortener) Shorten(url string) string {
	hash := md5.Sum([]byte(url))
	shortenURL := hex.EncodeToString(hash[:])[:8]
	s.storage.Set(shortenURL, url)
	return shortenURL
}

func (s *Shortener) GetOriginal(key string) (string, bool) {
	return s.storage.Get(key)
}
