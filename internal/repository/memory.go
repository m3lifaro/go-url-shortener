package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/m3lifaro/go-url-shortener/internal/model"
	"go.uber.org/zap"
)

type Storage interface {
	Get(key, userID string) (original string, existed bool, isDeleted bool, error error)
	GetAll(ctx context.Context, userID string) ([]model.UserLinkDto, error)
	Set(key, url, userID string) (string, error)
	Close() error
	BatchSet(records map[string]string, userID string) error
	BatchDelete(records []string, userID string) error
}

type MemoryStorage struct {
	mu        sync.RWMutex
	cache     map[string]model.ShortenRecord
	userIndex map[string]map[string]struct{}
	nextID    int
	producer  *Producer
	logger    *zap.Logger
}

func NewMemoryStorage(fileName string, logger *zap.Logger) (Storage, error) {
	consumer, err := NewConsumer(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create file consumer: %w", err)
	}
	defer consumer.Close()

	events, err := consumer.ReadAllEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to read events from file(%s): %w", fileName, err)
	}
	maxID := 0
	cache := make(map[string]model.ShortenRecord)
	linkMap := make(map[string]map[string]struct{})

	for _, rec := range *events {
		if r, ok := cache[rec.ShortenURL]; !ok || rec.CreatedAt.After(r.CreatedAt) {
			cache[rec.ShortenURL] = rec
		}
	}

	for _, event := range cache {
		if !event.IsDeleted {
			userMap, ok := linkMap[event.UserID]
			if !ok {
				userMap = make(map[string]struct{})
			}
			userMap[event.ShortenURL] = struct{}{}
			linkMap[event.UserID] = userMap
		}
		convertedID, err := strconv.Atoi(event.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ID: %w", err)
		}
		if convertedID > maxID {
			maxID = convertedID
		}
	}

	producer, err := NewProducer(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}
	return &MemoryStorage{
		cache:     cache,
		userIndex: linkMap,
		nextID:    maxID + 1,
		producer:  producer,
		logger:    logger,
	}, nil
}

func (s *MemoryStorage) Get(key, userID string) (original string, existed bool, isDeleted bool, error error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.cache[key]
	return val.URL, ok, val.IsDeleted, nil
}

func (s *MemoryStorage) GetAll(ctx context.Context, userID string) ([]model.UserLinkDto, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userCache, ok := s.userIndex[userID]
	if !ok {
		return []model.UserLinkDto{}, nil
	}
	response := make([]model.UserLinkDto, 0, len(userCache))
	for k := range userCache {
		val, ok := s.cache[k]
		if ok {
			response = append(response, model.UserLinkDto{
				OriginalURL: val.URL,
				ShortURL:    k,
			})
		}
	}
	return response, nil
}

func (s *MemoryStorage) Set(key, value, userID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var record model.ShortenRecord

	if val, exists := s.cache[key]; exists {
		return val.ShortenURL, nil
	} else {
		record = model.ShortenRecord{
			ID:         strconv.Itoa(s.nextID),
			URL:        value,
			ShortenURL: key,
			UserID:     userID,
			IsDeleted:  false,
			CreatedAt:  time.Now()}
		s.cache[key] = record
		userMap, ok := s.userIndex[userID]
		if !ok {
			userMap = make(map[string]struct{})
		}
		userMap[key] = struct{}{}
		s.userIndex[userID] = userMap
	}

	if err := s.producer.WriteEvent(&record); err != nil {
		return "", fmt.Errorf("failed to write event: %w", err)
	}

	s.nextID++
	return "", nil
}

func (s *MemoryStorage) BatchSet(records map[string]string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.userIndex[userID]
	if !ok {
		userCache := make(map[string]struct{})
		s.userIndex[userID] = userCache
	}
	for key := range records {
		if _, exists := s.cache[key]; exists {
			return fmt.Errorf("key %s already exists", key)
		}
	}

	var events []*model.ShortenRecord
	for key, value := range records {
		events = append(events, &model.ShortenRecord{
			ID:         strconv.Itoa(s.nextID),
			URL:        value,
			ShortenURL: key,
			UserID:     userID,
			IsDeleted:  false,
			CreatedAt:  time.Now()})
		s.nextID++
	}

	if err := s.producer.WriteEvents(events); err != nil {
		return fmt.Errorf("batch write failed: %w", err)
	}

	for _, value := range events {
		s.cache[value.ShortenURL] = *value
	}

	return nil
}

func (s *MemoryStorage) BatchDelete(records []string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userCache, ok := s.userIndex[userID]
	if !ok {
		return nil
	}

	for _, key := range records {
		delete(s.cache, key)
		delete(userCache, key)
	}

	s.userIndex[userID] = userCache

	return nil
}

func (s *MemoryStorage) Close() error {
	return s.producer.Close()
}

type Producer struct {
	file   *os.File
	writer *bufio.Writer
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) WriteEvent(event *model.ShortenRecord) error {
	data, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	return p.writer.Flush()
}

func (p *Producer) WriteEvents(events []*model.ShortenRecord) error {
	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		if _, err := p.writer.Write(data); err != nil {
			return fmt.Errorf("failed to write event data: %w", err)
		}

		if err := p.writer.WriteByte('\n'); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	if err := p.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	return nil
}

type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadEvent() (*model.ShortenRecord, error) {
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}
	data := c.scanner.Bytes()

	event := model.ShortenRecord{}
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (c *Consumer) ReadAllEvents() (*[]model.ShortenRecord, error) {
	var events []model.ShortenRecord
	for c.scanner.Scan() {
		data := c.scanner.Bytes()

		event := model.ShortenRecord{}
		err := json.Unmarshal(data, &event)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return &events, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

func (p *Producer) Close() error {
	return p.file.Close()
}
