package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jonesrussell/goprowl/search/crawlers"
	"github.com/jonesrussell/goprowl/search/engine"
	"go.uber.org/fx"
)

// Config holds application configuration
type Config struct {
	StartURL string
	MaxDepth int
}

// Application represents our running app
type Application struct {
	crawler    crawlers.Crawler
	engine     engine.SearchEngine
	config     *Config
	shutdowner fx.Shutdowner
}

// NewApplication creates a new Application instance
func NewApplication(
	crawler crawlers.Crawler,
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

// Search performs a search operation
func (app *Application) Search(queryStr string) error {
	processor := engine.NewQueryProcessor()
	query, err := processor.ParseQuery(queryStr)
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

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

// ListDocuments lists all indexed documents
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

// Shutdown gracefully shuts down the application
func (app *Application) Shutdown() {
	if err := app.shutdowner.Shutdown(); err != nil {
		log.Printf("Error shutting down: %v", err)
	}
}

// Run starts the application
func (app *Application) Run(ctx context.Context) error {
	// Create a context with timeout
	crawlCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Use the config's StartURL with context
	if err := app.crawler.Crawl(crawlCtx, app.config.StartURL); err != nil {
		return fmt.Errorf("crawl failed: %w", err)
	}
	return nil
}
