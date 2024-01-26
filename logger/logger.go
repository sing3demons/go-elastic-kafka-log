package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	*zap.Logger
}

type LoggerFields map[string]any

type ILogger interface {
	Info(msg string, fields ...LoggerFields)
	Error(msg string, fields ...LoggerFields)
	Debug(msg string, fields ...LoggerFields)

	Sync() error
	With(fields ...zapcore.Field) ILogger
}

func NewLogger() ILogger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
		InitialFields: map[string]interface{}{
			"pid": os.Getpid(),
		},
	}

	return &logger{zap.Must(config.Build())}
}

func (l *logger) Sync() error {
	return l.Logger.Sync()
}

func (l *logger) Info(msg string, fields ...LoggerFields) {
	var f []zapcore.Field
	for _, v := range fields {
		for key, value := range v {
			f = append(f, zap.Any(key, value))
		}
	}
	l.Logger.Info(msg, f...)
}

func (l *logger) Error(msg string, fields ...LoggerFields) {
	var f []zapcore.Field
	for _, v := range fields {
		for key, value := range v {
			f = append(f, zap.Any(key, value))
		}
	}
	l.Logger.Error(msg, f...)
}

func (l *logger) With(fields ...zapcore.Field) ILogger {
	return &logger{l.Logger.With(fields...)}
}

func (l *logger) Debug(msg string, fields ...LoggerFields) {
	var f []zapcore.Field
	for _, v := range fields {
		for key, value := range v {
			f = append(f, zap.Any(key, value))
		}
	}
	l.Logger.Debug(msg, f...)
}