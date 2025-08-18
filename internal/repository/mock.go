package repository

type MockStorage struct {
	GetFunc      func(key string) (string, bool, error)
	SetFunc      func(key, url string) (string, error)
	BatchSetFunc func(records map[string]string) error
}

func (m *MockStorage) BatchSet(records map[string]string) error {
	return m.BatchSetFunc(records)
}

func (m *MockStorage) Close() error {
	return nil
}

func (m *MockStorage) Get(key string) (string, bool, error) {
	return m.GetFunc(key)
}

func (m *MockStorage) Set(key, url string) (string, error) {
	return m.SetFunc(key, url)
}
