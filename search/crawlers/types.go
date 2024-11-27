package crawlers

import (
	"context"
	"time"
)

// Crawler defines the interface for web crawling operations
type Crawler interface {
	// Crawl starts crawling from the given URL up to specified depth
	Crawl(ctx context.Context, startURL string, depth int) error
	GetID() string
	CrawlWithHandler(ctx context.Context, startURL string, depth int, handler PageHandler) error
}

// CrawlResult represents the result of a crawl operation
type CrawlResult struct {
	URL       string
	Title     string
	Content   string
	Links     []string
	CreatedAt string
}

type CrawlConfig struct {
	MaxDepth int
}

type CrawlStatus struct {
	PagesVisited    int
	PagesSuccessful int
	PagesFailed     int
	CurrentDepth    int
	CurrentURL      string
	StartTime       time.Time
	LastUpdateTime  time.Time
	State           string // "running", "paused", "completed", "failed"
}

// Add to search/crawlers/types.go
type PageContent struct {
	URL         string
	Content     string
	ContentHash string
	LastUpdated time.Time
}

// PageHandler defines the callback for processing crawled pages
type PageHandler func(context.Context, *CrawlResult) error
