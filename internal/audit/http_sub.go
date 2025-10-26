package audit

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type HTTPSubscriber struct {
	url string
}

func NewHTTPSubscriber(url string) *HTTPSubscriber {
	return &HTTPSubscriber{url: url}
}

func (h *HTTPSubscriber) Notify(event Event) {
	data, _ := json.Marshal(event)
	resp, err := http.Post(h.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("cannot send audit: %v", err)
	}
	if resp != nil {
		resp.Body.Close()
	}
}
