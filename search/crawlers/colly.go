package crawlers

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/gocolly/colly/v2"
)

type CollyCrawler struct {
	collector *colly.Collector
}

func NewCollyCrawler(maxDepth int) Crawler {
	config := NewConfig()
	config.MaxDepth = maxDepth
	return NewCrawlerFromConfig(config)
}

// Crawl implements the Crawler interface
func (c *CollyCrawler) Crawl(ctx context.Context, startURL string, depth int) error {
	fmt.Printf("\nStarting crawl of %s with depth %d\n", startURL, depth)

	// Verify URL is valid
	parsedURL, err := url.Parse(startURL)
	if err != nil {
		return fmt.Errorf("invalid URL %s: %w", startURL, err)
	}

	// Allow the domain we're crawling
	c.collector.AllowedDomains = []string{parsedURL.Host}
	fmt.Printf("Set allowed domain: %s\n", parsedURL.Host)

	// Configure parallel requests
	c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 4,
		RandomDelay: 2 * time.Second,
	})

	fmt.Println("Starting initial visit...")
	err = c.collector.Visit(startURL)
	if err != nil {
		return fmt.Errorf("failed to start crawl of %s: %w", startURL, err)
	}

	// Add timeout for Wait()
	done := make(chan bool)
	go func() {
		c.collector.Wait()
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("Crawl completed successfully!")
	case <-time.After(2 * time.Minute):
		return fmt.Errorf("crawl timed out after 2 minutes")
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
