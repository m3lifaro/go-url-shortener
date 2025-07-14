package handler

import (
	"github.com/m3lifaro/go-url-shortener/internal/repository"
	"github.com/m3lifaro/go-url-shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path, contentType string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", contentType)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
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
	var rHandler = NewRedirectHandler(shortenService)
	var sHandler = NewShortenHandler(shortenService)
	ts := httptest.NewServer(NewRouter(&Handlers{
		Redirect: rHandler.ServeHTTP,
		Shorten:  sHandler.ServeHTTP,
	}))
	defer ts.Close()
	tests := []struct {
		method         string
		url            string
		expectedCode   int
		expectedBody   string
		expectedHeader []string
		body           io.Reader
	}{
		{method: http.MethodGet, url: "/ya", expectedCode: http.StatusOK},
		{method: http.MethodGet, url: "/not_found", expectedCode: http.StatusNotFound, expectedBody: "404 page not found\n"},
		{method: http.MethodPut, url: "/ya", expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodDelete, url: "/ya", expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPost, url: "/ya", expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "Empty url not allowed"},
		{method: http.MethodPost, expectedCode: http.StatusCreated, expectedBody: "http://localhost:8080/06509a58", body: strings.NewReader("ya.ru")},
	}
	for _, v := range tests {
		resp, body := testRequest(t, ts, v.method, v.url, "text/plain", v.body)
		resp.Body.Close()
		if v.expectedHeader != nil {
			assert.Equal(t, v.expectedHeader[1], resp.Header.Get(v.expectedHeader[0]), "Значение хидера не совпадает с ожидаемым")
		}
		assert.Equal(t, v.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
		if v.expectedBody != "" {
			assert.Equal(t, v.expectedBody, body, "Тело ответа не совпадает с ожидаемым")
		}
	}
}
