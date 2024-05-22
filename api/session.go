package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/project-n-oss/sidekick/app"
	"go.uber.org/zap"
)

type SessionContextKey string

const sessionContextKey SessionContextKey = "SessionContextKey"

func CtxSession(ctx context.Context) *app.Session {
	return ctx.Value(sessionContextKey).(*app.Session).WithContext(ctx)
}

type statusCodeRecorder struct {
	http.ResponseWriter
	http.Hijacker
	StatusCode int
}

func (r *statusCodeRecorder) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (api *Api) sessionMiddleware(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		beginTime := time.Now()
		session := api.app.NewSession()

		id := make([]byte, 20)
		if _, err := rand.Read(id); err != nil {
			session.Logger().Error("failed to generate random request id, setting default")
		}
		requestId := base64.RawURLEncoding.EncodeToString(id)
		session = session.WithLogger(session.Logger().With(
			zap.String("request_id", requestId),
			zap.String("user_agent", r.UserAgent()),
		))

		hijacker, _ := w.(http.Hijacker)
		w = &statusCodeRecorder{
			ResponseWriter: w,
			Hijacker:       hijacker,
		}

		defer func() {
			duration := time.Since(beginTime)
			method := r.Method
			host := r.Host
			path := r.URL.Path
			statusCode := w.(*statusCodeRecorder).StatusCode
			if statusCode == 0 {
				statusCode = 200
			}

			logger := session.Logger().With(
				zap.Duration("duration", duration),
				zap.String("method", method),
				zap.String("host", host),
				zap.String("path", path),
				zap.Int("statusCode", statusCode),
			)
			logger.Info(method + " " + path)
			if session.Logger().Level() == zap.DebugLevel {
				dump, err := httputil.DumpRequest(r, true)
				if err != nil {
					logger.Error("dumping session request", zap.Error(err))
					return
				}
				logger.Debug("session request dump", zap.String("dump", string(dump)))
			}
		}()

		newCtx := context.WithValue(r.Context(), sessionContextKey, session)
		r = r.WithContext(newCtx)

		handler.ServeHTTP(w, r)
	}
}
