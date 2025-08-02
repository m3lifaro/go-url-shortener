package repository

import (
	"bufio"
	"encoding/json"
	"github.com/m3lifaro/go-url-shortener/internal/model"
	"os"
	"strconv"
	"sync"
)

type Storage interface {
	Get(key string) (string, bool)
	Set(key, url string)
}

type MemoryStorage struct {
	mu       sync.RWMutex
	cache    map[string]string
	nextID   int
	fileName string
}

func NewMemoryStorage(fileName string) Storage {
	consumer, _ := NewConsumer(fileName)
	events, err := consumer.ReadAllEvents()
	if err != nil {
		panic(err)
	}
	maxID := 0
	for _, v := range *events {
		convertedID, err := strconv.Atoi(v.ID)
		if err != nil {
			panic(err)
		}
		if convertedID > maxID {
			maxID = convertedID
		}
	}
	defer consumer.Close()
	return &MemoryStorage{
		cache:    make(map[string]string),
		fileName: fileName,
		nextID:   maxID + 1,
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
	producer, _ := NewProducer(s.fileName)
	producer.WriteEvent(&model.ShortenRecord{ID: strconv.Itoa(s.nextID), URL: value, ShortenURL: key})
	s.nextID++
	defer s.mu.Unlock()
	s.cache[key] = value
}

type Producer struct {
	file *os.File
	// добавляем Writer в Producer
	writer *bufio.Writer
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file: file,
		// создаём новый Writer
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) WriteEvent(event *model.ShortenRecord) error {
	data, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return p.writer.Flush()
}

type Consumer struct {
	file *os.File
	// заменяем Reader на Scanner
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file: file,
		// создаём новый scanner
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadEvent() (*model.ShortenRecord, error) {
	// одиночное сканирование до следующей строки
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}
	// читаем данные из scanner
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
