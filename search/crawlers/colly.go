package crawlers

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
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
	// Only keep the lastVisit map for rate limiting
	lastVisit sync.Map
}

type Config struct {
	MaxDepth       int
	Parallelism    int
	RandomDelay    time.Duration
	AllowedHosts   []string
	FollowExternal bool
	CrawlTimeout   time.Duration
	// Updated configuration option name
	IgnoreRobots bool                     // Changed from RespectRobots
	RateLimit    map[string]time.Duration // Per-domain rate limits
	MaxRetries   int
	RetryDelay   time.Duration
}

func NewConfig() *Config {
	return &Config{
		MaxDepth:       3,
		Parallelism:    2,
		RandomDelay:    5 * time.Second,
		AllowedHosts:   []string{"*"},
		FollowExternal: false,
		CrawlTimeout:   30 * time.Second,
		IgnoreRobots:   false, // Default to respecting robots.txt
		RateLimit: map[string]time.Duration{
			"*": 1 * time.Second, // Default rate limit
		},
		MaxRetries: 3,
		RetryDelay: 5 * time.Second,
	}
}

// New creates a new CollyCrawler with dependencies injected by fx
func New(
	engine engine.SearchEngine,
	config *Config,
) *CollyCrawler {
	options := []colly.CollectorOption{
		colly.MaxDepth(config.MaxDepth),
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.82 Safari/537.36"),
	}

	// Add robots.txt configuration
	if config.IgnoreRobots {
		options = append(options, colly.IgnoreRobotsTxt())
	}

	c := colly.NewCollector(options...)

	// Add random user agent if respecting robots.txt
	if !config.IgnoreRobots {
		extensions.RandomUserAgent(c)
	}

	// Configure rate limiting
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: config.Parallelism,
		RandomDelay: config.RandomDelay,
	})

	crawler := &CollyCrawler{
		collector: c,
		engine:    engine,
		config:    config,
	}

	crawler.setupCallbacks()
	return crawler
}

func (c *CollyCrawler) setupCallbacks() {
	c.collector.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting: %v", r.URL)

		// Apply per-domain rate limiting
		domain := r.URL.Host
		if err := c.checkRateLimit(domain); err != nil {
			r.Abort()
			return
		}
	})

	c.collector.OnResponse(func(r *colly.Response) {
		log.Printf("Got response %d from %v", r.StatusCode, r.Request.URL)
	})

	c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		normalizedURL := c.normalizeURL(e.Request.URL, link)
		if normalizedURL != "" {
			e.Request.Visit(normalizedURL)
		}
	})

	c.collector.OnHTML("html", func(e *colly.HTMLElement) {
		title := e.ChildText("title")
		content := e.Text
		url := e.Request.URL.String()

		doc := NewDocument(url, title, content)
		if err := c.engine.Index(doc); err != nil {
			log.Printf("Error indexing document: %v", err)
		}
	})

	c.collector.OnError(func(r *colly.Response, err error) {
		log.Printf("Error visiting %v: %v", r.Request.URL, err)

		// Get the current retry count from context or default to 0
		retryCount := 0
		if count := r.Ctx.GetAny("retry_count"); count != nil {
			retryCount = count.(int)
		}

		if r.StatusCode >= 500 && retryCount < c.config.MaxRetries {
			retryCount++
			retryAfter := time.Duration(retryCount) * c.config.RetryDelay
			time.Sleep(retryAfter)

			// Store updated retry count
			r.Ctx.Put("retry_count", retryCount)

			// Retry the request
			r.Request.Visit(r.Request.URL.String())
		}
	})
}

func (c *CollyCrawler) checkRateLimit(domain string) error {
	now := time.Now()

	// Get domain-specific rate limit or use default
	rateLimit, ok := c.config.RateLimit[domain]
	if !ok {
		rateLimit = c.config.RateLimit["*"]
	}

	// Check and update last visit time
	if lastVisitI, ok := c.lastVisit.Load(domain); ok {
		lastVisit := lastVisitI.(time.Time)
		if now.Sub(lastVisit) < rateLimit {
			return fmt.Errorf("rate limit exceeded for domain: %s", domain)
		}
	}

	c.lastVisit.Store(domain, now)
	return nil
}

func (c *CollyCrawler) normalizeURL(baseURL *url.URL, rawURL string) string {
	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Handle relative URLs
	resolvedURL := baseURL.ResolveReference(parsedURL)

	// Clean the path
	resolvedURL.Path = strings.TrimSuffix(resolvedURL.Path, "/")
	resolvedURL.Path = strings.TrimSuffix(resolvedURL.Path, "index.html")

	// Remove common tracking parameters
	q := resolvedURL.Query()
	paramsToRemove := []string{"utm_source", "utm_medium", "utm_campaign", "fbclid", "gclid"}
	for _, param := range paramsToRemove {
		q.Del(param)
	}
	resolvedURL.RawQuery = q.Encode()

	// Remove fragments unless they're part of single-page applications
	resolvedURL.Fragment = ""

	return resolvedURL.String()
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
