package crawlers

import "context"

// Crawler defines the interface for web crawling operations
type Crawler interface {
	// Crawl starts crawling from the given URL up to specified depth
	Crawl(ctx context.Context, startURL string, depth int) error
}

// CrawlResult represents the result of a crawl operation
type CrawlResult struct {
	URL       string
	Title     string
	Content   string
	Links     []string
	CreatedAt string
}
