package repository

type MockStorage struct {
	GetFunc func(key string) (string, bool, error)
	SetFunc func(key, url string) error
}

func (m *MockStorage) Close() error {
	return nil
}

func (m *MockStorage) Get(key string) (string, bool, error) {
	return m.GetFunc(key)
}

func (m *MockStorage) Set(key, url string) error {
	return m.SetFunc(key, url)
}
