package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/m3lifaro/go-url-shortener/internal/model"
	"go.uber.org/zap"
)

type Storage interface {
	Get(key, userID string) (string, bool, error)
	GetAll(userID string) ([]model.UserResponseItem, error)
	Set(key, url, userID string) (string, error)
	Close() error
	BatchSet(records map[string]string, userID string) error
}

type MemoryStorage struct {
	mu       sync.RWMutex
	cache    map[string]map[string]string
	linkMap  map[string]string
	nextID   int
	producer *Producer
	logger   *zap.Logger
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
	cache := make(map[string]map[string]string)
	linkMap := make(map[string]string)
	for _, v := range *events {
		userCache, ok := cache[v.UserID]
		if !ok {
			userCache = make(map[string]string)
			cache[v.UserID] = userCache
		}
		if e, exists := userCache[v.ShortenURL]; exists {
			logger.Warn("got duplicated event for user",
				zap.String("user_id", v.UserID),
				zap.String("shorten_url", v.ShortenURL),
				zap.String("url", v.URL),
				zap.String("already_presented_as", e),
			)
			continue
		}
		userCache[v.ShortenURL] = v.URL
		linkMap[v.ShortenURL] = v.URL
		convertedID, err := strconv.Atoi(v.ID)
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
		cache:    cache,
		linkMap:  linkMap,
		nextID:   maxID + 1,
		producer: producer,
		logger:   logger,
	}, nil
}

func (s *MemoryStorage) Get(key, userID string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	//userCache, ok := s.cache[userID]
	//if !ok {
	//	return "", false, nil
	//}
	val, ok := s.linkMap[key]
	//val, ok := userCache[key]
	return val, ok, nil
}

func (s *MemoryStorage) GetAll(userID string) ([]model.UserResponseItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userCache, ok := s.cache[userID]
	if !ok {
		return []model.UserResponseItem{}, nil
	}
	response := make([]model.UserResponseItem, 0, len(userCache))
	for k, v := range userCache {
		response = append(response, model.UserResponseItem{
			OriginalURL: v,
			ShortURL:    k,
		})
	}
	return response, nil
}

func (s *MemoryStorage) Set(key, value, userID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	userCache, ok := s.cache[userID]
	if !ok {
		userCache = make(map[string]string)
		s.cache[userID] = userCache
	}
	if _, exists := userCache[key]; exists {
		return "", fmt.Errorf("key %s already exists", key)
	}

	for k, v := range userCache {
		if v == value {
			return k, nil
		}
	}

	record := &model.ShortenRecord{ID: strconv.Itoa(s.nextID), URL: value, ShortenURL: key, UserID: userID}

	if err := s.producer.WriteEvent(record); err != nil {
		return "", fmt.Errorf("failed to write event: %w", err)
	}

	userCache[key] = value
	s.nextID++
	return "", nil
}

func (s *MemoryStorage) BatchSet(records map[string]string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userCache, ok := s.cache[userID]
	if !ok {
		userCache = make(map[string]string)
		s.cache[userID] = userCache
	}
	for key := range records {
		if _, exists := userCache[key]; exists {
			return fmt.Errorf("key %s already exists", key)
		}
	}

	var events []*model.ShortenRecord
	for key, value := range records {
		events = append(events, &model.ShortenRecord{
			ID:         strconv.Itoa(s.nextID),
			URL:        value,
			ShortenURL: key,
		})
		s.nextID++
	}

	if err := s.producer.WriteEvents(events); err != nil {
		return fmt.Errorf("batch write failed: %w", err)
	}

	for key, value := range records {
		userCache[key] = value
	}

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
