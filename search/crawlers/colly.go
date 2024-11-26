package crawlers

import (
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
}

func NewConfig() *Config {
	return &Config{
		MaxDepth:       3,
		Parallelism:    2,
		RandomDelay:    5 * time.Second,
		AllowedHosts:   []string{"*"},
		FollowExternal: false,
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

	return &CollyCrawler{
		collector: c,
		engine:    engine,
		config:    config,
	}
}

func (c *CollyCrawler) Crawl(url string) error {
	log.Printf("Starting crawl of: %s", url)

	// Parse the starting URL to get the domain
	startDomain, err := extractDomain(url)
	if err != nil {
		return fmt.Errorf("invalid start URL: %v", err)
	}
	log.Printf("Base domain: %s", startDomain)

	// Allow external domains if configured
	if c.config.FollowExternal {
		c.collector.AllowedDomains = nil // Allow all domains
		log.Printf("Following external domains enabled")
	} else {
		c.collector.AllowedDomains = []string{startDomain}
		log.Printf("Restricting to domain: %s", startDomain)
	}

	// Set up callbacks
	c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Only follow links that are absolute or can be made absolute
		absoluteURL := e.Request.AbsoluteURL(link)
		if absoluteURL != "" {
			log.Printf("Following link: %s", absoluteURL)
			err := e.Request.Visit(absoluteURL)
			if err != nil {
				log.Printf("Error visiting %s: %v", absoluteURL, err)
			}
		}
	})

	c.collector.OnHTML("html", func(e *colly.HTMLElement) {
		// Extract title from meta tags or title tag
		title := e.ChildText("title")
		if title == "" {
			title = e.ChildAttr("meta[property='og:title']", "content")
		}
		if title == "" {
			title = e.Request.URL.String() // Use URL as fallback title
		}

		// Extract content from body
		content := e.ChildText("body")

		// Create document with both title and content
		doc := NewDocument(e.Request.URL.String(), title, content)

		// Index the document
		if err := c.engine.Index(doc); err != nil {
			log.Printf("Error indexing page %s: %v", e.Request.URL.String(), err)
		}
	})

	c.collector.OnError(func(r *colly.Response, err error) {
		log.Printf("Error crawling %s: %v", r.Request.URL, err)
	})

	// Start the crawl
	err = c.collector.Visit(url)
	if err != nil {
		return err
	}

	c.collector.Wait()
	log.Printf("Crawl completed for: %s", url)

	return nil
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
