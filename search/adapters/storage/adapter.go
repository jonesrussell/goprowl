package storage

import (
	"context"
	"path/filepath"
	"time"

	"github.com/jonesrussell/goprowl/search/crawlers"
	"github.com/jonesrussell/goprowl/search/storage"
	"github.com/jonesrussell/goprowl/search/storage/bleve"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// StorageAdapter wraps the storage implementation
type StorageAdapter struct {
	storage storage.StorageAdapter
	logger  *zap.Logger
}

// NewStorageAdapter creates a new storage adapter
func NewStorageAdapter(logger *zap.Logger) (*StorageAdapter, error) {
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

// HandleCrawledPage stores a crawled page in the storage
func (a *StorageAdapter) HandleCrawledPage(ctx context.Context, result *crawlers.CrawlResult) error {
	a.logger.Debug("handling crawled page",
		zap.String("url", result.URL),
		zap.String("title", result.Title),
		zap.Int("content_length", len(result.Content)))

	doc := &storage.Document{
		URL:       result.URL,
		Title:     result.Title,
		Content:   result.Content,
		Type:      "webpage",
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"links":      result.Links,
			"created_at": result.CreatedAt,
		},
	}

	if err := a.storage.Store(ctx, doc); err != nil {
		a.logger.Error("failed to store document",
			zap.String("url", result.URL),
			zap.Error(err))
		return err
	}

	a.logger.Info("stored document successfully",
		zap.String("url", result.URL),
		zap.String("title", result.Title),
		zap.Int("links_count", len(result.Links)))
	return nil
}
