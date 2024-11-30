package app

import (
	"github.com/jonesrussell/goprowl/search/core/types"
	"github.com/jonesrussell/goprowl/search/engine"
	"github.com/jonesrussell/goprowl/search/storage"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// DefaultConfig provides default application configuration
func DefaultConfig() *Config {
	return &Config{
		StartURL: "",
		MaxDepth: 3,
	}
}

// Module combines all application dependencies
var Module = fx.Options(
	StorageModule,
	EngineModule,
	fx.Provide(
		DefaultConfig,
		NewApplication,
	),
)

// StorageModule provides storage dependencies
var StorageModule = fx.Options(
	fx.Provide(
		fx.Annotate(
			func(logger *zap.Logger) (types.StorageAdapter, error) {
				return storage.NewStorageAdapter(logger)
			},
			fx.As(new(types.StorageAdapter)),
		),
	),
)

// EngineModule provides search engine dependencies
var EngineModule = fx.Options(
	fx.Provide(engine.New),
)
