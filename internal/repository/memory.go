package repository

import (
	"fmt"
	"sync"
)

type MemoryStorage struct {
	mu    sync.RWMutex
	cache map[string]string
}

func NewMemoryStorage() *MemoryStorage {
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
	fmt.Println("BOMBAKLAT")
	for k, v := range s.cache {
		fmt.Println(k, v)
	}
}
