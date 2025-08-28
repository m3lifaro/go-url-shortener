package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/m3lifaro/go-url-shortener/internal/model"
	"github.com/m3lifaro/go-url-shortener/internal/repository"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRedirectHandler_ServeHTTP(t *testing.T) {
	mock := &repository.MockStorage{
		SetFunc: func(key, url, userID string) (string, error) {
			return "", nil
		},
		GetFunc: func(key, userID string) (string, bool, error) {
			if key == "not_found" {
				return "", false, nil
			}
			return "https://ya.ru", true, nil
		},
		GetAllFunc: func(userID string) ([]model.UserResponseItem, error) {
			return make([]model.UserResponseItem, 0), nil
		},
	}

	var zl = zap.NewNop()
	var shortenService = service.NewShortener(mock)
	var handler = NewRedirectHandler(shortenService, zl)
	testCases := []struct {
		method         string
		url            string
		expectedCode   int
		expectedBody   string
		expectedHeader string
	}{
		{method: http.MethodGet, url: "ya", expectedCode: http.StatusTemporaryRedirect, expectedBody: "", expectedHeader: "https://ya.ru"},
		{method: http.MethodGet, url: "not_found", expectedCode: http.StatusNotFound, expectedBody: ""},
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
