package repository

import (
	"sync"
)

type Storage interface {
	Get(key string) (string, bool)
	Set(key, url string)
}

type MemoryStorage struct {
	mu    sync.RWMutex
	cache map[string]string
}

func NewMemoryStorage() Storage {
	return &MemoryStorage{
		cache: make(map[string]string),
	}
}

func (s *MemoryStorage) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.cache[key]
	return val, ok
}

func (s *MemoryStorage) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[key] = value
}
