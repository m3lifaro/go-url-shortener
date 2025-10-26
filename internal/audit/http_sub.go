package audit

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type HttpSubscriber struct {
	url string
}

func NewHttpSubscriber(url string) *HttpSubscriber {
	return &HttpSubscriber{url: url}
}

func (h *HttpSubscriber) Notify(event Event) {
	data, _ := json.Marshal(event)
	resp, err := http.Post(h.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("cannot send audit: %v", err)
	}
	if resp != nil {
		resp.Body.Close()
	}
}
