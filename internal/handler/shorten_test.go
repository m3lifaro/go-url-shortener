package handler

import (
	"github.com/m3lifaro/go-url-shortener/internal/repository"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"path"
	"regexp"
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
	var handler = NewShortenHandler(shortenService, "http://localhost:8080/")
	var validHeader = "text/plain; charset=utf-8"
	var invalidHeader = "application/json; charset=utf-8"
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
		{method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "Unsupported Content-Type. Expected 'text/plain', got: application/json", header: invalidHeader},
		{method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "Empty url not allowed", header: validHeader},
		{method: http.MethodPost, expectedCode: http.StatusCreated, header: validHeader, body: "ya.ru"},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))
			r.Header.Set("Content-Type", tc.header)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if tc.method == http.MethodPost && tc.body != "" {
				hashed := path.Base(w.Body.String())
				validURLPattern := regexp.MustCompile(`^[A-Za-z0-9_-]+$`).MatchString
				assert.Equal(t, true, validURLPattern(hashed))
				assert.Equal(t, 30, len(w.Body.String()))
			} else {
				assert.Equal(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}
