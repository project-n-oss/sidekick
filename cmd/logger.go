package cmd

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"
)

func NewLogger(logLevel zapcore.Level) *zap.Logger {
	var logEncoder zapcore.Encoder
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		encoderConfig := getEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		logEncoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		logEncoder = zapcore.NewConsoleEncoder(getEncoderConfig())
	}

	zapAtom := zap.NewAtomicLevel()
	zapAtom.SetLevel(logLevel)

	ret := zap.New(
		zapcore.NewCore(
			logEncoder,
			zapcore.Lock(os.Stdout),
			zapAtom,
		),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)

	zapAtom.SetLevel(logLevel)

	return ret
}

func getEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = customMilliTimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.CallerKey = "caller"
	return encoderConfig
}

func customMilliTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02:15:04:05.000"))
}
