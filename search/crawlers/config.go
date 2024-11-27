package crawlers

import (
	"github.com/gocolly/colly/v2"
)

// Config holds crawler configuration
type Config struct {
	MaxDepth       int
	AllowedDomains []string
	UserAgent      string
}

// NewConfig creates a default crawler configuration
func NewConfig() *Config {
	return &Config{
		MaxDepth:       3,
		AllowedDomains: []string{},
		UserAgent:      "GoProwl Bot",
	}
}

// NewCrawlerFromConfig creates a new CollyCrawler from configuration
func NewCrawlerFromConfig(config *Config) *CollyCrawler {
	c := colly.NewCollector(
		colly.MaxDepth(config.MaxDepth),
		colly.UserAgent(config.UserAgent),
	)

	if len(config.AllowedDomains) > 0 {
		c.AllowedDomains = config.AllowedDomains
	}

	return &CollyCrawler{
		collector: c,
		maxDepth:  config.MaxDepth,
	}
}
