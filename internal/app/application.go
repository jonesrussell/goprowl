package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jonesrussell/goprowl/internal/logger"
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
	logger     logger.Logger
}

// NewApplication creates a new Application instance
func NewApplication(
	crawler crawlers.Crawler,
	engine engine.SearchEngine,
	config *Config,
	shutdowner fx.Shutdowner,
	logger logger.Logger,
) *Application {
	return &Application{
		crawler:    crawler,
		engine:     engine,
		config:     config,
		shutdowner: shutdowner,
		logger:     logger,
	}
}

// Search performs a search operation
func (app *Application) Search(ctx context.Context, queryStr string) error {
	processor := engine.NewQueryProcessor()
	query, err := processor.ParseQuery(queryStr)
	if err != nil {
		app.logger.Error(ctx, "failed to parse query", app.logger.NewField("query", queryStr))
		return fmt.Errorf("failed to parse query: %w", err)
	}

	results, err := app.engine.Search(query)
	if err != nil {
		app.logger.Error(ctx, "search failed", app.logger.NewField("query", queryStr))
		return fmt.Errorf("search failed: %w", err)
	}

	app.logger.Info(ctx, "search completed", app.logger.NewField("total_results", results.Metadata["total"].(int64)))
	total := results.Metadata["total"].(int64)
	fmt.Printf("Found %d results:\n\n", total)
	for _, hit := range results.Hits {
		content := hit.Content
		fmt.Printf("Title: %s\n", content["title"])
		fmt.Printf("URL: %s\n", content["url"])
		fmt.Printf("Type: %s\n", content["type"])
		fmt.Println("---")
	}

	return nil
}

// ListDocuments lists all indexed documents with proper error handling and metrics
func (app *Application) ListDocuments(ctx context.Context) error {
	app.logger.Info(ctx, "retrieving document list")

	docs, err := app.engine.List()
	if err != nil {
		app.logger.Error(ctx, "failed to list documents", app.logger.NewField("error", err))
		return fmt.Errorf("failed to list documents: %w", err)
	}

	app.logger.Info(ctx, "documents retrieved successfully", app.logger.NewField("document_count", len(docs)))

	return nil
}

// Shutdown gracefully shuts down the application
func (app *Application) Shutdown(ctx context.Context) {
	if err := app.shutdowner.Shutdown(); err != nil {
		app.logger.Error(ctx, "error shutting down", app.logger.NewField("error", err))
	}
	app.logger.Info(ctx, "application shutdown complete")
}

// Run starts the application
func (app *Application) Run(ctx context.Context) error {
	app.logger.Info(ctx, "starting application", app.logger.NewField("start_url", app.config.StartURL), app.logger.NewField("max_depth", app.config.MaxDepth))

	crawlCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if err := app.engine.Clear(); err != nil {
		app.logger.Error(ctx, "failed to clear existing data", app.logger.NewField("error", err))
		return fmt.Errorf("failed to clear existing data: %w", err)
	}

	if err := app.crawler.Crawl(crawlCtx, app.config.StartURL, app.config.MaxDepth); err != nil {
		app.logger.Error(ctx, "crawl failed", app.logger.NewField("error", err))
		return fmt.Errorf("crawl failed: %w", err)
	}

	app.logger.Info(ctx, "application run completed successfully")
	return nil
}
