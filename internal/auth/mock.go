package auth

type MockAuth struct {
	ParseJWTFunc    func(token string) (*UserClaims, error)
	GenerateJWTFunc func(userID string) (string, error)
}

func (m *MockAuth) ParseJWT(tokenString string) (*UserClaims, error) {
	return m.ParseJWTFunc(tokenString)
}
func (m *MockAuth) GenerateJWT(userID string) (string, error) {
	return m.GenerateJWTFunc(userID)
}
