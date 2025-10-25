package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/m3lifaro/go-url-shortener/internal/auth"
	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			uri := r.RequestURI
			method := r.Method
			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}
			next.ServeHTTP(&lw, r)

			duration := time.Since(start)

			logger.Info("got incoming HTTP request",
				zap.String("method", method),
				zap.String("uri", uri),
				zap.Duration("duration", duration),
				zap.Int("status", responseData.status),
				zap.Int("size", responseData.size),
			)
		}

		return http.HandlerFunc(logFn)
	}
}

func gzipMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				cw := newCompressWriter(w)
				ow = cw
				defer cw.Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					logger.Error("got error creating compress reader",
						zap.Error(err),
					)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer cr.Close()
			}

			next.ServeHTTP(ow, r)
		})
	}
}

func authMiddleware(logger *zap.Logger, auz auth.Auth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID string
			var hasAuth bool

			userID, hasAuth = parseUserCookie(r, auz, logger)
			if hasAuth && userID == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if !hasAuth {
				userID = uuid.New().String()
			}

			ctx := createAuthContext(r, userID, hasAuth)

			rw := generateTokenAndWriter(w, userID, hasAuth, auz, logger)
			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}

func generateTokenAndWriter(w http.ResponseWriter, userID string, hasAuth bool, auz auth.Auth, logger *zap.Logger) *responseWriter {
	token, err := auz.GenerateJWT(userID)

	if err == nil {
		logger.Debug("got auth response",
			zap.String("user_id", userID))
	}
	if err != nil {
		logger.Error("got auth error",
			zap.Error(err))
	}
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK, token: token, shouldSetCookie: !hasAuth}
	return rw
}

func createAuthContext(r *http.Request, userID string, hasAuth bool) context.Context {
	ctx := context.WithValue(r.Context(), auth.UserIDKey, userID)
	ctx = context.WithValue(ctx, auth.HasAuthKey, hasAuth)
	ctx = context.WithValue(ctx, auth.ShouldSetCookieKey, !hasAuth)
	return ctx
}

func parseUserCookie(r *http.Request, auz auth.Auth, logger *zap.Logger) (userID string, hasAuth bool) {
	cookie, err := r.Cookie(auth.CookieName)
	if err == nil {
		claims, err := auz.ParseJWT(cookie.Value)
		if err == nil {
			userID = claims.UUID
			hasAuth = true
			return userID, hasAuth
		} else {
			logger.Debug("got error parsing cookie", zap.Error(err), zap.String("cookie", cookie.Value))
		}
	}
	return userID, false
}

type responseWriter struct {
	http.ResponseWriter
	statusCode      int
	cookie          *http.Cookie
	shouldSetCookie bool
	token           string
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	if rw.shouldSetCookie && (rw.statusCode < 400 || rw.statusCode == 409) {
		http.SetCookie(rw, &http.Cookie{
			Name:  auth.CookieName,
			Value: rw.token,
		})
	}

	rw.ResponseWriter.WriteHeader(code)
}
