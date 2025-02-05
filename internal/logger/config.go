// internal/logger/config.go
package logger

import (
	"os"

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

// LoadConfigFromEnv loads logger configuration from environment variables
func LoadConfigFromEnv() Config {
	return Config{
		Level:            zap.NewAtomicLevelAt(getLogLevel()),
		OutputPaths:      []string{os.Getenv("LOG_OUTPUT_PATHS")},
		ErrorOutputPaths: []string{os.Getenv("LOG_ERROR_OUTPUT_PATHS")},
	}
}

// getLogLevel retrieves log level from environment variable
func getLogLevel() zapcore.Level {
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}
