package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/jonesrussell/goprowl/search/core/types"
	"github.com/jonesrussell/goprowl/search/crawlers"
	"github.com/jonesrussell/goprowl/search/storage/bleve"
	"github.com/jonesrussell/goprowl/search/storage/document"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// StorageAdapter wraps the storage implementation
type StorageAdapter struct {
	storage types.StorageAdapter
	logger  *zap.Logger
}

// NewStorageAdapter creates a new storage adapter
func NewStorageAdapter(logger *zap.Logger) (types.StorageAdapter, error) {
	storagePath := filepath.Join("data", "search.bleve")
	storage, err := bleve.New(storagePath)
	if err != nil {
		return nil, err
	}

	logger.Info("initialized storage adapter", zap.String("path", storagePath))
	return &StorageAdapter{
		storage: storage,
		logger:  logger,
	}, nil
}

// Module provides storage adapter dependencies
var Module = fx.Module("storage",
	fx.Provide(NewStorageAdapter),
)

// Store implements the StorageAdapter interface
func (a *StorageAdapter) Store(ctx context.Context, doc types.Document) error {
	return a.storage.Store(ctx, doc)
}

// BatchStore implements the StorageAdapter interface
func (a *StorageAdapter) BatchStore(ctx context.Context, docs []types.Document) error {
	return a.storage.BatchStore(ctx, docs)
}

// HandleCrawledPage stores a crawled page in the storage
func (a *StorageAdapter) HandleCrawledPage(ctx context.Context, result *crawlers.CrawlResult) error {
	logger := a.logger.With(
		zap.String("operation", "handle_crawled_page"),
		zap.String("url", result.URL),
		zap.String("title", result.Title),
		zap.Int("content_length", len(result.Content)),
		zap.String("crawl_time", result.CreatedAt),
	)

	logger.Debug("processing crawled page")

	doc := document.NewDocument(
		result.URL,
		result.Title,
		result.Content,
		"webpage",
		time.Now(),
		map[string]interface{}{
			"links":      result.Links,
			"created_at": result.CreatedAt,
		},
	)

	if err := a.Store(ctx, doc); err != nil {
		logger.Error("failed to store document",
			zap.String("url", result.URL),
			zap.Error(err))
		return err
	}

	logger.Info("stored document successfully",
		zap.String("url", result.URL),
		zap.String("title", result.Title),
		zap.Int("links_count", len(result.Links)))
	return nil
}

// Get implements the StorageAdapter interface
func (a *StorageAdapter) Get(ctx context.Context, id string) (types.Document, error) {
	return a.storage.Get(ctx, id)
}

// GetAll implements the StorageAdapter interface
func (a *StorageAdapter) GetAll(ctx context.Context) ([]types.Document, error) {
	return a.storage.GetAll(ctx)
}

// Delete implements the StorageAdapter interface
func (a *StorageAdapter) Delete(ctx context.Context, id string) error {
	return a.storage.Delete(ctx, id)
}

// Clear implements the StorageAdapter interface
func (a *StorageAdapter) Clear(ctx context.Context) error {
	return a.storage.Clear(ctx)
}

// Close implements the StorageAdapter interface
func (a *StorageAdapter) Close() error {
	if err := a.storage.Close(); err != nil {
		a.logger.Error("failed to close storage", zap.Error(err))
		return fmt.Errorf("storage close failed: %w", err)
	}
	a.logger.Info("storage closed successfully")
	return nil
}
