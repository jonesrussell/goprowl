package engine

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Add structured configuration
type Config struct {
	MaxResults     int           `json:"max_results" validate:"required,min=1,max=1000"`
	SearchTimeout  time.Duration `json:"search_timeout" validate:"required"`
	CacheEnabled   bool          `json:"cache_enabled"`
	IndexBatchSize int           `json:"index_batch_size" validate:"required,min=1"`
}

func ValidateConfig(cfg *Config) error {
	return validator.New().Struct(cfg)
}
