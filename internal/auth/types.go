package auth

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UUID string `json:"uuid"`
	jwt.RegisteredClaims
}

type contextKey string

const (
	UserIDKey          contextKey = "user_id"
	HasAuthKey         contextKey = "has_auth"
	ShouldSetCookieKey contextKey = "should_set_cookie"
	CookieName         string     = "authorization"
)

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
