package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jonesrussell/goprowl/search/crawlers"
	"github.com/jonesrussell/goprowl/search/engine"
	"github.com/jonesrussell/goprowl/search/storage"
	"github.com/jonesrussell/goprowl/search/storage/memory"
	"go.uber.org/fx"
)

// Module options for different components
var StorageModule = fx.Options(
	fx.Provide(
		memory.New,
		fx.Annotate(
			memory.New,
			fx.As(new(storage.StorageAdapter)),
		),
	),
)

var EngineModule = fx.Options(
	fx.Provide(engine.New),
)

var CrawlerModule = fx.Options(
	fx.Provide(
		crawlers.New,
		NewCrawlerConfig,
	),
)

// Application configuration
type Config struct {
	StartURL string
	MaxDepth int
}

func NewConfig() *Config {
	startURL := flag.String("url", "https://go.dev", "The URL to start crawling from")
	maxDepth := flag.Int("depth", 3, "Maximum crawl depth")
	flag.Parse()

	return &Config{
		StartURL: *startURL,
		MaxDepth: *maxDepth,
	}
}

// Application represents our running app
type Application struct {
	crawler *crawlers.CollyCrawler
	engine  engine.SearchEngine
	config  *Config
}

func NewApplication(
	crawler *crawlers.CollyCrawler,
	engine engine.SearchEngine,
	config *Config,
) *Application {
	return &Application{
		crawler: crawler,
		engine:  engine,
		config:  config,
	}
}

// Run starts the application
func (app *Application) Run(ctx context.Context) error {
	// Use the config's StartURL instead of hardcoding
	if err := app.crawler.Crawl(app.config.StartURL); err != nil {
		return err
	}
	return nil
}

func NewCrawlerConfig(config *Config) *crawlers.Config {
	return &crawlers.Config{
		MaxDepth:    config.MaxDepth,
		Parallelism: 2,
		RandomDelay: 1 * time.Second,
	}
}

func (app *Application) Search(queryString string) error {
	query := engine.NewBasicQuery(strings.Fields(queryString))

	results, err := app.engine.Search(query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	total := results.Metadata["total"].(int64)
	fmt.Printf("Found %d results:\n\n", total)
	for _, hit := range results.Hits {
		content := hit.Content()
		fmt.Printf("Title: %s\n", content["title"])
		fmt.Printf("URL: %s\n", content["url"])
		fmt.Printf("Type: %s\n", hit.Type())
		fmt.Println("---")
	}

	return nil
}

func main() {
	app := fx.New(
		// Provide all dependencies
		fx.Provide(
			NewConfig,
			NewApplication,
		),

		// Include our modules
		StorageModule,
		EngineModule,
		CrawlerModule,

		// Start the application
		fx.Invoke(func(app *Application) {
			if err := app.Run(context.Background()); err != nil {
				log.Fatal(err)
			}
		}),
	)

	// Start the application
	app.Run()
}
