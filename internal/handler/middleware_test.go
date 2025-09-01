package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/m3lifaro/go-url-shortener/internal/auth"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAuthMiddlewareOptional(t *testing.T) {
	//logger := zaptest.NewLogger(t)
	lvl, _ := zap.ParseAtomicLevel("DEBUG")
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	logger, _ := cfg.Build()
	t.Run("Valid cookie", func(t *testing.T) {
		auz := &auth.MockAuth{
			ParseJWTFunc: func(token string) (*auth.UserClaims, error) {
				return &auth.UserClaims{UUID: "user-123"}, nil
			},
			GenerateJWTFunc: func(userID string) (string, error) {
				return "token-123", nil
			},
		}
		mw := authMiddlewareOptional(logger, auz)

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
		assert.Empty(t, w.Result().Cookies()) // Cookie не добавляется
	})

	t.Run("No cookie", func(t *testing.T) {
		auz := &auth.MockAuth{
			ParseJWTFunc: func(token string) (*auth.UserClaims, error) {
				t.Fatal("ParseJWT должен не вызываться")
				return nil, nil
			},
			GenerateJWTFunc: func(userID string) (string, error) {
				return "gen-token", nil
			},
		}
		mw := authMiddlewareOptional(logger, auz)

		nextCalled := false
		h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID := ctx.Value(auth.UserIDKey).(string)
			hasAuth := ctx.Value(auth.HasAuthKey).(bool)
			shouldSetCookie := ctx.Value(auth.ShouldSetCookieKey).(bool)

			_, err := uuid.Parse(userID)
			assert.NoError(t, err) // Это сгенерированный UUID
			assert.False(t, hasAuth)
			assert.True(t, shouldSetCookie)
			nextCalled = true
		}))

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.True(t, nextCalled)
		assert.Equal(t, http.StatusOK, w.Code)
		cookies := w.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, "gen-token", cookies[0].Value)
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
		mw := authMiddlewareOptional(logger, auz)

		h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Хендлер не должен вызываться")
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: "bad-uuid"})
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
