package boltrouter

import "go.uber.org/zap"

// LeveledLogger is used as a wrapper around zap.Logger to implement the retryablehttp.LeveledLogger interface.
// This allow zap.Logger to work with retryablehttp.Client
type LeveledLogger struct {
	logger *zap.Logger
}

func NewLeveledLogger(logger *zap.Logger) LeveledLogger {
	return LeveledLogger{
		logger: logger,
	}
}

func (l LeveledLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Error(msg, zap.Any("retryableHttp", keysAndValues))
}

func (l LeveledLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, zap.Any("retryableHttp", keysAndValues))

}

func (l LeveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debug(msg, zap.Any("retryableHttp", keysAndValues))

}

func (l LeveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warn(msg, zap.Any("retryableHttp", keysAndValues))
}
