package app

import (
	"github.com/jonesrussell/goprowl/search/engine"
	"github.com/jonesrussell/goprowl/search/storage"
	"github.com/jonesrussell/goprowl/search/storage/bleve"
	"go.uber.org/fx"
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
			func() (storage.StorageAdapter, error) {
				return bleve.New("data/search.bleve")
			},
			fx.As(new(storage.StorageAdapter)),
		),
	),
)

// EngineModule provides search engine dependencies
var EngineModule = fx.Options(
	fx.Provide(engine.New),
)
