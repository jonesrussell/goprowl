package crawlers

import (
	"log"
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
	MaxDepth     int
	Parallelism  int
	RandomDelay  time.Duration
	AllowedHosts []string
}

func NewConfig() *Config {
	return &Config{
		MaxDepth:     3,
		Parallelism:  2,
		RandomDelay:  5 * time.Second,
		AllowedHosts: []string{"*"},
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
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: config.Parallelism,
		RandomDelay: config.RandomDelay,
	})

	return &CollyCrawler{
		collector: c,
		engine:    engine,
		config:    config,
	}
}

func (c *CollyCrawler) Crawl(url string) error {
	log.Printf("Starting crawl of: %s", url)

	// Set up callbacks
	c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.Printf("Found link: %s", link)
	})

	c.collector.OnHTML("body", func(e *colly.HTMLElement) {
		log.Printf("Visiting page: %s", e.Request.URL)
		// Extract content and store it
		content := e.Text
		err := c.engine.Index(NewDocument(e.Request.URL.String(), content))
		if err != nil {
			log.Printf("Error indexing page %s: %v", e.Request.URL, err)
		}
	})

	c.collector.OnError(func(r *colly.Response, err error) {
		log.Printf("Error crawling %s: %v", r.Request.URL, err)
	})

	// Start the crawl
	return c.collector.Visit(url)
}

// Create a document type for crawled pages
type CrawledDocument struct {
	url     string
	content map[string]interface{}
}

func NewDocument(url string, content string) engine.Document {
	return &CrawledDocument{
		url: url,
		content: map[string]interface{}{
			"url":     url,
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
