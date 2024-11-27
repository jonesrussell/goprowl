package crawlers

import (
	"context"

	"github.com/gocolly/colly/v2"
)

// CollyCrawler implements web crawling using colly
type CollyCrawler struct {
	collector *colly.Collector
	maxDepth  int
}

// NewCollyCrawler creates a new Crawler implementation
func NewCollyCrawler(maxDepth int) Crawler {
	c := colly.NewCollector(
		colly.MaxDepth(maxDepth),
	)

	return &CollyCrawler{
		collector: c,
		maxDepth:  maxDepth,
	}
}

// Crawl implements the crawling logic
func (c *CollyCrawler) Crawl(ctx context.Context, startURL string) error {
	// Set up context handling
	go func() {
		<-ctx.Done()
		c.collector.AllowedDomains = []string{} // This effectively stops new requests
	}()

	err := c.collector.Visit(startURL)
	if err != nil {
		return err
	}

	// Wait for all crawling goroutines to finish
	c.collector.Wait()
	return nil
}
