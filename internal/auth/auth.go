package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type Auth interface {
	GenerateJWT(userID string) (string, error)
	ParseJWT(tokenString string) (*UserClaims, error)
}
type AuthImpl struct {
	secret []byte
}

func NewAuth(secret string) *AuthImpl {
	return &AuthImpl{secret: []byte(secret)}
}

func (a *AuthImpl) GenerateJWT(userID string) (string, error) {
	claims := &UserClaims{
		UUID: userID, //rn without any other claims
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secret)
}

func (a *AuthImpl) ParseJWT(tokenString string) (*UserClaims, error) {
	claims := &UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return a.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
