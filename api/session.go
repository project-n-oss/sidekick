package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/project-n-oss/sidekick/boltrouter"
	"go.uber.org/zap"
)

type Session struct {
	br      *boltrouter.BoltRouter
	logger  *zap.Logger
	context context.Context
}

func (a *Api) NewSession() *Session {
	return &Session{
		br:     a.br,
		logger: a.logger,
	}
}

func (s *Session) WithLogger(logger *zap.Logger) *Session {
	s.logger = logger
	return s
}
func (s *Session) Logger() *zap.Logger {
	return s.logger
}

func (s *Session) WithContext(ctx context.Context) *Session {
	s.context = ctx
	return s
}

type SessionContextKey string

const sessionContextKey SessionContextKey = "SessionContextKey"

func CtxSession(ctx context.Context) *Session {
	return ctx.Value(sessionContextKey).(*Session).WithContext(ctx)
}

func (a *Api) sessionMiddleware(handler http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		beginTime := time.Now()
		session := a.NewSession()

		id := make([]byte, 20)
		if _, err := rand.Read(id); err != nil {
			a.logger.Error("failed to generate random request id, setting default")
		}
		requestId := base64.RawURLEncoding.EncodeToString(id)
		session = session.WithLogger(session.Logger().With(
			zap.String("request_id", requestId),
			zap.String("user_agent", r.UserAgent()),
		))

		defer func() {
			duration := time.Since(beginTime)
			method := r.Method
			host := r.Host
			path := r.URL.Path
			logger := session.Logger().With(
				zap.Duration("duration", duration),
				zap.String("method", method),
				zap.String("host", host),
				zap.String("path", path),
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
	})
}
