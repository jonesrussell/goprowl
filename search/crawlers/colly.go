package crawlers

import (
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

func (c *CollyCrawler) Crawl(startURL string) error {
	// Handle page content
	c.collector.OnHTML("body", func(e *colly.HTMLElement) {
		// Create a document for the search engine
		doc := engine.NewDocument(
			e.Request.URL.String(), // ID
			"webpage",              // Type
		)

		// Set content
		doc.SetContent("title", e.ChildText("title"))
		doc.SetContent("body", e.Text)
		doc.SetContent("url", e.Request.URL.String())

		// Set metadata
		doc.SetMetadata("crawled_at", time.Now())
		doc.SetMetadata("domain", e.Request.URL.Host)

		// Index the document
		c.engine.Index(doc)
	})

	return c.collector.Visit(startURL)
}
