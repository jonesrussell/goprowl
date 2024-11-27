package crawlers

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/goprowl/metrics"
)

type CollyCrawler struct {
	collector *colly.Collector
	metrics   *metrics.ComponentMetrics
	id        string
}

func NewCollyCrawler(collector *colly.Collector, metrics *metrics.ComponentMetrics) *CollyCrawler {
	uniqueID := fmt.Sprintf("crawler-%d", time.Now().UnixNano())

	return &CollyCrawler{
		collector: collector,
		metrics:   metrics,
		id:        uniqueID,
	}
}

func setupCallbacks(c *colly.Collector, m *metrics.ComponentMetrics) {
	c.OnRequest(func(r *colly.Request) {
		m.IncrementActiveRequests()
		fmt.Printf("Starting request to: %s\n", r.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		m.DecrementActiveRequests()
		m.IncrementErrorCount()
		fmt.Printf("Error visiting %s: %v\n", r.Request.URL, err)
	})

	c.OnResponse(func(r *colly.Response) {
		m.DecrementActiveRequests()
		m.IncrementPagesProcessed()
		m.ObserveResponseSize(float64(len(r.Body)))
		fmt.Printf("Got response from %s\n", r.Request.URL)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absLink := e.Request.AbsoluteURL(link)
		if absLink != "" {
			fmt.Printf("Found link: %s\n", absLink)
			e.Request.Visit(absLink)
		}
	})
}

// GetID returns the crawler's unique identifier
func (c *CollyCrawler) GetID() string {
	return c.id
}

// Crawl implements the Crawler interface
func (c *CollyCrawler) Crawl(ctx context.Context, startURL string, depth int) error {
	fmt.Printf("Crawler %s starting crawl of %s with depth %d\n", c.id, startURL, depth)

	// Set up callbacks before starting the crawl
	setupCallbacks(c.collector, c.metrics)

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
