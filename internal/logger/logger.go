package logger

import (
	"context"

	"go.uber.org/zap"
)

// Logger is an interface for logging
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)
}

// Field represents a log field
type Field = zap.Field

// zapLogger ensures implementation of the Logger interface
type zapLogger struct {
	logger *zap.Logger
}

func (l *zapLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.logger.Debug(msg, fields...)
}

func (l *zapLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.logger.Info(msg, fields...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.logger.Warn(msg, fields...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.logger.Error(msg, fields...)
}

func (l *zapLogger) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.logger.Fatal(msg, fields...)
}
