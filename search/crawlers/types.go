package crawlers

import "context"

// Crawler defines the interface for web crawling functionality
type Crawler interface {
	Crawl(ctx context.Context, startURL string) error
}
