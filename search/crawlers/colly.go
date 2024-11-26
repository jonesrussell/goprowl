package crawlers

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/goprowl/search/engine"
	"go.uber.org/fx"
)

// Module provides fx options for the crawler
var Module = fx.Options(
	fx.Provide(New),
	fx.Provide(NewConfig),
)

type CollyCrawler struct {
	collector *colly.Collector
	engine    engine.SearchEngine
	config    *Config
}

type Config struct {
	MaxDepth       int
	Parallelism    int
	RandomDelay    time.Duration
	AllowedHosts   []string
	FollowExternal bool
	CrawlTimeout   time.Duration
}

func NewConfig() *Config {
	return &Config{
		MaxDepth:       3,
		Parallelism:    2,
		RandomDelay:    5 * time.Second,
		AllowedHosts:   []string{"*"},
		FollowExternal: false,
		CrawlTimeout:   30 * time.Second,
	}
}

// New creates a new CollyCrawler with dependencies injected by fx
func New(
	engine engine.SearchEngine,
	config *Config,
) *CollyCrawler {
	c := colly.NewCollector(
		colly.MaxDepth(config.MaxDepth),
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.82 Safari/537.36"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: config.Parallelism,
		RandomDelay: config.RandomDelay,
	})

	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting: %v", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("Got response %d from %v", r.StatusCode, r.Request.URL)
	})

	// Add handlers for links and content
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		e.Request.Visit(link)
	})

	c.OnHTML("html", func(e *colly.HTMLElement) {
		// Create a document from the page
		title := e.ChildText("title")
		content := e.Text
		url := e.Request.URL.String()

		doc := NewDocument(url, title, content)

		// Index the document
		if err := engine.Index(doc); err != nil {
			log.Printf("Error indexing document: %v", err)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error visiting %v: %v", r.Request.URL, err)
	})

	return &CollyCrawler{
		collector: c,
		engine:    engine,
		config:    config,
	}
}

func (c *CollyCrawler) Crawl(ctx context.Context, urlStr string) error {
	log.Printf("Starting crawl of: %s", urlStr)

	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Ensure URL has scheme
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	// Get the normalized URL string
	normalizedURL := parsedURL.String()

	// Parse the starting URL to get the domain
	startDomain, err := extractDomain(normalizedURL)
	if err != nil {
		return fmt.Errorf("invalid start URL: %w", err)
	}
	log.Printf("Base domain: %s", startDomain)

	// Set allowed domains
	if c.config.FollowExternal {
		c.collector.AllowedDomains = nil
	} else {
		c.collector.AllowedDomains = []string{startDomain}
		log.Printf("Restricting to domain: %s", startDomain)
	}

	// Create a done channel for graceful shutdown
	done := make(chan bool)
	var crawlErr error

	go func() {
		defer close(done)

		// Use Visit directly with error handling
		if err := c.collector.Visit(normalizedURL); err != nil {
			crawlErr = fmt.Errorf("visit failed: %w", err)
			return
		}

		c.collector.Wait()
	}()

	// Wait for either context cancellation or crawl completion
	select {
	case <-ctx.Done():
		return fmt.Errorf("crawl interrupted: %w", ctx.Err())
	case <-done:
		if crawlErr != nil {
			return fmt.Errorf("crawl error: %w", crawlErr)
		}
		log.Printf("Crawl completed for: %s", normalizedURL)
		return nil
	}
}

// Helper function to extract domain from URL
func extractDomain(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	return parsedURL.Host, nil
}

// Create a document type for crawled pages
type CrawledDocument struct {
	url     string
	content map[string]interface{}
}

func NewDocument(url string, title string, content string) engine.Document {
	return &CrawledDocument{
		url: url,
		content: map[string]interface{}{
			"url":     url,
			"title":   title,
			"content": content,
		},
	}
}

// Implement the engine.Document interface
func (d *CrawledDocument) ID() string                       { return d.url }
func (d *CrawledDocument) Type() string                     { return "page" }
func (d *CrawledDocument) Content() map[string]interface{}  { return d.content }
func (d *CrawledDocument) Metadata() map[string]interface{} { return map[string]interface{}{} }
func (d *CrawledDocument) Permission() *engine.Permission   { return &engine.Permission{} }
