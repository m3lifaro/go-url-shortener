package repository

type MockStorage struct {
	GetFunc func(key string) (string, bool)
	SetFunc func(key, url string)
}

func (m *MockStorage) Get(key string) (string, bool) {
	return m.GetFunc(key)
}

func (m *MockStorage) Set(key, url string) {
	m.SetFunc(key, url)
}
