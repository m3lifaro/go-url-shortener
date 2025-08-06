package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/m3lifaro/go-url-shortener/internal/model"
	"go.uber.org/zap"
	"os"
	"strconv"
	"sync"
)

type Storage interface {
	Get(key string) (string, bool, error)
	Set(key, url string) error
	Close() error
}

type MemoryStorage struct {
	mu       sync.RWMutex
	cache    map[string]string
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
	cache := make(map[string]string)
	for _, v := range *events {
		if e, exists := cache[v.ShortenURL]; exists {
			logger.Warn("got duplicated event",
				zap.String("shorten_url", v.ShortenURL),
				zap.String("url", v.URL),
				zap.String("already_presented_as", e),
			)
			continue
		}
		cache[v.ShortenURL] = v.URL
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
		cache:    make(map[string]string),
		nextID:   maxID + 1,
		producer: producer,
		logger:   logger,
	}, nil
}

func (s *MemoryStorage) Get(key string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.cache[key]
	return val, ok, nil
}

func (s *MemoryStorage) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.cache[key]; exists {
		return fmt.Errorf("key %s already exists", key)
	}

	record := &model.ShortenRecord{ID: strconv.Itoa(s.nextID), URL: value, ShortenURL: key}

	if err := s.producer.WriteEvent(record); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	s.cache[key] = value
	s.nextID++
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
