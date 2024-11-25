package crawler

import (
	"time"

	"github.com/jonesrussell/goprowl/search/engine"

	"github.com/gocolly/colly/v2"
)

type CollyCrawler struct {
	collector *colly.Collector
	engine    engine.SearchEngine
}

func New(searchEngine engine.SearchEngine) *CollyCrawler {
	c := colly.NewCollector(
		colly.MaxDepth(3),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		RandomDelay: 5 * time.Second,
	})

	return &CollyCrawler{
		collector: c,
		engine:    searchEngine,
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
