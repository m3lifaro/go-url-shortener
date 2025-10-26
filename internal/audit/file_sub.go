package audit

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

type FileSubscriber struct {
	file *os.File
	mu   sync.Mutex
}

func NewFileSubscriber(path string) *FileSubscriber {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("can't open audit log file: %v", err)
	}
	return &FileSubscriber{file: file}
}

func (f *FileSubscriber) Notify(event Event) {
	data, _ := json.Marshal(event)

	f.mu.Lock()
	defer f.mu.Unlock()

	f.file.Write(append(data, '\n'))
}
