package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/m3lifaro/go-url-shortener/internal/repository"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirectHandler_ServeHTTP(t *testing.T) {
	mock := &repository.MockStorage{
		SetFunc: func(key, url string) {
		},
		GetFunc: func(key string) (string, bool) {
			if key == "not_found" {
				return "", false
			}
			return "https://ya.ru", true
		},
	}

	var shortenService = service.NewShortener(mock)
	var handler = NewRedirectHandler(shortenService)
	testCases := []struct {
		method         string
		url            string
		expectedCode   int
		expectedBody   string
		expectedHeader string
	}{
		{method: http.MethodGet, url: "ya", expectedCode: http.StatusTemporaryRedirect, expectedBody: "", expectedHeader: "https://ya.ru"},
		{method: http.MethodGet, url: "not_found", expectedCode: http.StatusNotFound, expectedBody: "404 page not found\n"},
		{method: http.MethodPut, url: "ya", expectedCode: http.StatusMethodNotAllowed, expectedBody: ""},
		{method: http.MethodDelete, url: "ya", expectedCode: http.StatusMethodNotAllowed, expectedBody: ""},
		{method: http.MethodPost, url: "ya", expectedCode: http.StatusMethodNotAllowed, expectedBody: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/"+tc.url, nil)
			w := httptest.NewRecorder()
			chi := chi.NewRouter()
			chi.Get("/{id}", handler.ServeHTTP)
			chi.ServeHTTP(w, r)
			if r := w.Header().Get("Location"); r != "" {
				assert.Equal(t, tc.expectedHeader, r)
			}
			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
		})
	}
}
