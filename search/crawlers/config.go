package crawlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
)

// ConfigOptions holds command-line parameters for the crawler
type ConfigOptions struct {
	URL      string
	MaxDepth int
}

// Config holds crawler configuration
type Config struct {
	ConfigOptions  // Embed the options
	AllowedDomains []string
	UserAgent      string
}

// ProvideDefaultConfigOptions creates default options
func ProvideDefaultConfigOptions() *ConfigOptions {
	return &ConfigOptions{
		MaxDepth: 1,
		URL:      "",
	}
}

// NewConfig creates a crawler configuration from options
func NewConfig(opts *ConfigOptions) *Config {
	return &Config{
		ConfigOptions:  *opts,
		AllowedDomains: []string{},
		UserAgent:      "GoProwl Bot",
	}
}

// NewCrawlerFromConfig creates a new CollyCrawler from configuration
func NewCrawlerFromConfig(config *Config) *CollyCrawler {
	c := colly.NewCollector(
		colly.MaxDepth(config.MaxDepth),
		colly.Async(true),
		colly.UserAgent(config.UserAgent),
		colly.Debugger(&debug.LogDebugger{}),
	)

	fmt.Printf("Configured crawler with MaxDepth: %d\n", config.MaxDepth)

	if len(config.AllowedDomains) > 0 {
		c.AllowedDomains = config.AllowedDomains
	}

	// Configure transport
	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
		MaxIdleConns:      10,
		IdleConnTimeout:   30 * time.Second,
	})

	// Set timeouts
	c.SetRequestTimeout(30 * time.Second)

	// Set up callbacks
	setupCallbacks(c)

	return &CollyCrawler{
		collector: c,
	}
}

// setupCallbacks configures all the collector callbacks
func setupCallbacks(c *colly.Collector) {
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("Starting request to: %s\n", r.URL)
		fmt.Printf("Request headers: %v\n", r.Headers)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error visiting %s: %v\n", r.Request.URL, err)
		log.Printf("Response status: %d\n", r.StatusCode)
		log.Printf("Response headers: %v\n", r.Headers)
		if len(r.Body) > 0 {
			log.Printf("Response body (first 200 bytes): %s\n", r.Body[:min(200, len(r.Body))])
		}
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Printf("Got response from %s (Status: %d)\n", r.Request.URL, r.StatusCode)
		fmt.Printf("Content-Type: %s\n", r.Headers.Get("Content-Type"))
		fmt.Printf("Body length: %d bytes\n", len(r.Body))
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absLink := e.Request.AbsoluteURL(link)
		if absLink != "" {
			fmt.Printf("Found link at depth %d: %s\n", e.Request.Depth, absLink)
			e.Request.Visit(absLink)
		}
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		fmt.Printf("Processing page: %s\n", e.Request.URL)
		// Here you can extract and process page content
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Printf("Finished scraping: %s\n", r.Request.URL)
		fmt.Printf("Crawling statistics:\n")
		fmt.Printf("  - Current URL depth: %d\n", r.Request.Depth)
		fmt.Printf("  - Response status: %d\n", r.StatusCode)
		fmt.Printf("  - Response size: %d bytes\n", len(r.Body))
	})
}
