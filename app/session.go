package app

import (
	"context"

	"go.uber.org/zap"
)

type Session struct {
	app     *App
	logger  *zap.Logger
	context context.Context
}

func (a *App) NewSession() *Session {
	return &Session{
		logger:  a.logger,
		app:     a,
		context: context.Background(),
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

func (s *Session) Context() context.Context {
	return s.context
}
