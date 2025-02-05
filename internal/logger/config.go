package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds the logger configuration
type Config struct {
	Level            zap.AtomicLevel
	OutputPaths      []string
	ErrorOutputPaths []string
}

// Build creates a new logger instance based on the Config
func (c Config) Build() (*zap.Logger, error) {
	cfg := zap.Config{
		Level:            c.Level,
		OutputPaths:      c.OutputPaths,
		ErrorOutputPaths: c.ErrorOutputPaths,
		EncoderConfig:    zap.NewProductionEncoderConfig(),
	}
	return cfg.Build()
}

// New creates a new logger instance
func New(config Config) (Logger, error) {
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return &zapLogger{logger: logger}, nil
}

// NewField creates a new log field
func NewField(key string, value interface{}) Field {
	return zap.Any(key, value)
}

// NewProductionConfig creates a new production logger configuration
func NewProductionConfig() Config {
	return Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// NewDevelopmentConfig creates a new development logger configuration
func NewDevelopmentConfig() Config {
	return Config{
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// NewAtomicLevelAt creates a new atomic level for logging
func NewAtomicLevelAt(level zapcore.Level) zap.AtomicLevel {
	return zap.NewAtomicLevelAt(level)
}
