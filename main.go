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
	"github.com/jonesrussell/goprowl/search/storage/bleve"
	"go.uber.org/fx"
)

// Module options for different components
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
	crawler    *crawlers.CollyCrawler
	engine     engine.SearchEngine
	config     *Config
	shutdowner fx.Shutdowner
}

func NewApplication(
	crawler *crawlers.CollyCrawler,
	engine engine.SearchEngine,
	config *Config,
	shutdowner fx.Shutdowner,
) *Application {
	return &Application{
		crawler:    crawler,
		engine:     engine,
		config:     config,
		shutdowner: shutdowner,
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
	listCmd := flag.Bool("list", false, "List all indexed documents")
	searchQuery := flag.String("search", "", "Search indexed documents")

	app := fx.New(
		StorageModule,
		EngineModule,
		CrawlerModule,
		fx.Provide(NewConfig),
		fx.Provide(NewApplication),
		fx.Invoke(func(app *Application, lifecycle fx.Lifecycle) {
			lifecycle.Append(fx.Hook{
				OnStart: func(context.Context) error {
					if *listCmd {
						if err := app.ListDocuments(); err != nil {
							log.Printf("Error listing documents: %v", err)
						}
						// Signal to stop the application
						go func() {
							app.Shutdown()
						}()
						return nil
					}

					if *searchQuery != "" {
						if err := app.Search(*searchQuery); err != nil {
							log.Printf("Error searching documents: %v", err)
						}
						// Signal to stop the application
						go func() {
							app.Shutdown()
						}()
						return nil
					}

					// For crawling
					if err := app.Run(context.Background()); err != nil {
						log.Printf("Error running application: %v", err)
					}
					// Signal to stop after crawling
					go func() {
						app.Shutdown()
					}()
					return nil
				},
			})
		}),
	)

	app.Run()
}

// Add this method to Application struct
func (app *Application) ListDocuments() error {
	docs, err := app.engine.List()
	if err != nil {
		return fmt.Errorf("failed to list documents: %w", err)
	}

	fmt.Printf("Found %d documents:\n\n", len(docs))
	for _, doc := range docs {
		content := doc.Content()
		fmt.Printf("Title: %s\n", content["title"])
		fmt.Printf("URL: %s\n", content["url"])
		fmt.Printf("Type: %s\n", doc.Type())
		fmt.Printf("Created: %s\n", doc.Metadata()["created_at"])
		fmt.Println("---")
	}

	return nil
}

func (app *Application) Shutdown() {
	if err := app.shutdowner.Shutdown(); err != nil {
		log.Printf("Error shutting down: %v", err)
	}
}
