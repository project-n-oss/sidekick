package cmd

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(verbose bool) *zap.Logger {
	logEncoder := getConsoleEncoder()

	zapAtom := zap.NewAtomicLevel()
	zapAtom.SetLevel(zapcore.InfoLevel)

	ret := zap.New(
		zapcore.NewCore(
			logEncoder,
			zapcore.Lock(os.Stdout),
			zapAtom,
		),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	zapAtom.SetLevel(zapcore.InfoLevel)

	if verbose {
		zapAtom.SetLevel(zapcore.DebugLevel)
	}

	OnShutdown(func() {
		_ = logger.Sync()
	})

	return ret
}

func getConsoleEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = customMilliTimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.CallerKey = "caller"
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func customMilliTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02:15:04:05.000"))
}
