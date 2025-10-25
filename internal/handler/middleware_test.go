package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m3lifaro/go-url-shortener/internal/auth"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestAuthMiddlewareOptional(t *testing.T) {
	logger := zaptest.NewLogger(t)
	t.Run("Valid cookie", func(t *testing.T) {
		auz := &auth.MockAuth{
			ParseJWTFunc: func(token string) (*auth.UserClaims, error) {
				return &auth.UserClaims{UUID: "user-123"}, nil
			},
			GenerateJWTFunc: func(userID string) (string, error) {
				return "token-123", nil
			},
		}
		mw := authMiddleware(logger, auz)

		nextCalled := false
		h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID := ctx.Value(auth.UserIDKey).(string)
			hasAuth := ctx.Value(auth.HasAuthKey).(bool)
			shouldSetCookie := ctx.Value(auth.ShouldSetCookieKey).(bool)

			assert.Equal(t, "user-123", userID)
			assert.True(t, hasAuth)
			assert.False(t, shouldSetCookie)
			nextCalled = true
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: "good"})
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.True(t, nextCalled)
		assert.Equal(t, http.StatusOK, w.Code)
		response := w.Result()
		assert.Empty(t, response.Cookies())
		_ = response.Body.Close()
	})

	t.Run("Claims UUID is empty", func(t *testing.T) {
		auz := &auth.MockAuth{
			ParseJWTFunc: func(token string) (*auth.UserClaims, error) {
				return &auth.UserClaims{UUID: ""}, nil
			},
			GenerateJWTFunc: func(userID string) (string, error) {
				t.Fatal("GenerateJWT должен не вызываться")
				return "", nil
			},
		}
		mw := authMiddleware(logger, auz)

		h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Handler shouldn't be called")
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: "bad-uuid"})
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
