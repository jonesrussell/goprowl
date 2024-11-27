package app

import (
	"github.com/jonesrussell/goprowl/search/crawlers"
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
	CrawlerModule,
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

// CrawlerModule provides crawler dependencies
var CrawlerModule = fx.Options(
	fx.Provide(
		crawlers.NewConfig,
		fx.Annotate(
			crawlers.NewCrawlerFromConfig,
			fx.As(new(crawlers.Crawler)),
		),
	),
)
