package repository

import "github.com/m3lifaro/go-url-shortener/internal/model"

type MockStorage struct {
	GetFunc      func(key, userID string) (string, bool, error)
	GetAllFunc   func(userID string) ([]model.UserLinkDto, error)
	SetFunc      func(key, url, userID string) (string, error)
	BatchSetFunc func(records map[string]string, userID string) error
}

func (m *MockStorage) BatchSet(records map[string]string, userID string) error {
	return m.BatchSetFunc(records, userID)
}

func (m *MockStorage) Close() error {
	return nil
}

func (m *MockStorage) Get(key, userID string) (string, bool, error) {
	return m.GetFunc(key, userID)
}
func (m *MockStorage) GetAll(userID string) ([]model.UserLinkDto, error) {
	return m.GetAllFunc(userID)
}

func (m *MockStorage) Set(key, url, userID string) (string, error) {
	return m.SetFunc(key, url, userID)
}
