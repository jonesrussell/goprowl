package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jonesrussell/goprowl/search/crawlers"
	"github.com/jonesrussell/goprowl/search/engine"
	"github.com/jonesrussell/goprowl/search/engine/query"
	"go.uber.org/fx"
	"go.uber.org/zap"
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
	logger     *zap.Logger
}

// NewApplication creates a new Application instance
func NewApplication(
	crawler crawlers.Crawler,
	engine engine.SearchEngine,
	config *Config,
	shutdowner fx.Shutdowner,
	logger *zap.Logger,
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
func (app *Application) Search(queryStr string) error {
	processor := query.NewQueryProcessor()
	q, err := processor.ParseQuery(queryStr)
	if err != nil {
		app.logger.Error("failed to parse query",
			zap.String("query", queryStr),
			zap.Error(err))
		return fmt.Errorf("failed to parse query: %w", err)
	}

	results, err := app.engine.Search(q)
	if err != nil {
		app.logger.Error("search failed",
			zap.String("query", queryStr),
			zap.Error(err))
		return fmt.Errorf("search failed: %w", err)
	}

	app.logger.Info("search completed",
		zap.Int64("total_results", results.Metadata["total"].(int64)))
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
func (app *Application) ListDocuments() error {
	app.logger.Info("retrieving document list")

	docs, err := app.engine.List()
	if err != nil {
		app.logger.Error("failed to list documents", zap.Error(err))
		return fmt.Errorf("failed to list documents: %w", err)
	}

	app.logger.Info("documents retrieved successfully",
		zap.Int("document_count", len(docs)),
	)

	return nil
}

// Shutdown gracefully shuts down the application
func (app *Application) Shutdown() {
	if err := app.shutdowner.Shutdown(); err != nil {
		app.logger.Error("error shutting down", zap.Error(err))
	}
	app.logger.Info("application shutdown complete")
}

// Run starts the application
func (app *Application) Run(ctx context.Context) error {
	app.logger.Info("starting application",
		zap.String("start_url", app.config.StartURL),
		zap.Int("max_depth", app.config.MaxDepth))

	crawlCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if err := app.engine.Clear(); err != nil {
		app.logger.Error("failed to clear existing data", zap.Error(err))
		return fmt.Errorf("failed to clear existing data: %w", err)
	}

	if err := app.crawler.Crawl(crawlCtx, app.config.StartURL, app.config.MaxDepth); err != nil {
		app.logger.Error("crawl failed", zap.Error(err))
		return fmt.Errorf("crawl failed: %w", err)
	}

	app.logger.Info("application run completed successfully")
	return nil
}
