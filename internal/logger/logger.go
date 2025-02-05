// internal/logger/logger.go
package logger

import (
	"context"

	"go.uber.org/zap"
)

// Logger is the interface for logging
type Logger interface {
	Info(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)
	NewField(key string, value interface{}) Field
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// zapLogger ensures implementation of the Logger interface
type zapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new zapLogger instance.
func NewZapLogger(z *zap.Logger) Logger {
	return &zapLogger{logger: z}
}

// Implement the Logger interface methods
func (l *zapLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.logger.Info(msg, convertFields(fields)...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.logger.Error(msg, convertFields(fields)...)
}

func (l *zapLogger) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.logger.Fatal(msg, convertFields(fields)...)
}

func (l *zapLogger) NewField(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// convertFields converts custom fields to zap fields
func convertFields(fields []Field) []zap.Field {
	var zapFields []zap.Field
	for _, f := range fields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}
	return zapFields
}
