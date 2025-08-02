package handler

import (
	"compress/gzip"
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
	var respBody []byte
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	contentEncoding := resp.Header.Get("Content-Encoding")
	contentTyp := resp.Header.Get("Content-Type")

	if strings.Contains(contentEncoding, "gzip") || strings.Contains(contentTyp, "application/x-gzip") {
		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		respBody, err = io.ReadAll(zr)
		require.NoError(t, err)

	} else {
		respBody, err = io.ReadAll(resp.Body)
		require.NoError(t, err)
	}
	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	mock := &repository.MockStorage{
		SetFunc: func(key, url string) error {
			return nil
		},
		GetFunc: func(key string) (string, bool, error) {
			if key == "not_found" {
				return "", false, nil
			}
			return "https://ya.ru", true, nil
		},
	}

	var shortenService = service.NewShortener(mock)
	ts := httptest.NewServer(NewRouter(NewHandlers(shortenService, "http://localhost:8080/")))
	defer ts.Close()
	tests := []struct {
		method         string
		url            string
		expectedCode   int
		expectedBody   string
		expectedHeader []string
		body           io.Reader
		contentType    string
	}{
		{method: http.MethodGet, url: "/ya", expectedCode: http.StatusOK},
		{method: http.MethodGet, url: "/not_found", expectedCode: http.StatusNotFound},
		{method: http.MethodPut, url: "/ya", expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodDelete, url: "/ya", expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPost, url: "/ya", expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "Empty url not allowed"},
		{method: http.MethodPost, expectedCode: http.StatusCreated, body: strings.NewReader("ya.ru"), expectedHeader: []string{"Content-Type", "text/plain"}},
		{method: http.MethodPost, url: "/api/shorten", expectedCode: http.StatusCreated, body: strings.NewReader(`{"url": "ya.ru"}`), contentType: "application/json", expectedHeader: []string{"Content-Type", "application/json"}},
	}
	for _, v := range tests {
		requestCT := "text/plain"
		if v.contentType != "" {
			requestCT = v.contentType
		}
		resp, body := testRequest(t, ts, v.method, v.url, requestCT, v.body)
		_ = resp.Body.Close()
		if v.expectedHeader != nil {
			assert.Equal(t, v.expectedHeader[1], resp.Header.Get(v.expectedHeader[0]), "Значение хидера не совпадает с ожидаемым")
		}
		assert.Equal(t, v.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
		if v.expectedBody != "" {
			assert.Equal(t, v.expectedBody, body, "Тело ответа не совпадает с ожидаемым")
		}
	}
}
