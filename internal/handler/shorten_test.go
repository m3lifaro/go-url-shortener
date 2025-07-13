package handler

import (
	"github.com/m3lifaro/go-url-shortener/internal/repository"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortenHandler_ServeHTTP(t *testing.T) {
	mock := &repository.MockStorage{
		SetFunc: func(key, url string) {
		},
		GetFunc: func(key string) (string, bool) {
			return "https://ya.ru", true
		},
	}

	var shortenService = service.NewShortener(mock)
	var handler = NewShortenHandler(shortenService)
	var validHeader = "text/plain"
	var invalidHeader = "application/json; charset=utf-8"
	// описываем набор данных: метод запроса, ожидаемый код ответа, ожидаемое тело
	testCases := []struct {
		method       string
		expectedCode int
		expectedBody string
		header       string
		body         string
	}{
		{method: http.MethodGet, expectedCode: http.StatusMethodNotAllowed, expectedBody: "", header: validHeader},
		{method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed, expectedBody: "", header: validHeader},
		{method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed, expectedBody: "", header: validHeader},
		{method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "Unsupported content-type. Only 'text/plain' allowed", header: invalidHeader},
		{method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "Empty url not allowed", header: validHeader},
		{method: http.MethodPost, expectedCode: http.StatusCreated, expectedBody: "http://localhost:8080/06509a58", header: validHeader, body: "ya.ru"},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))
			r.Header.Set("Content-Type", tc.header)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
		})
	}
}
