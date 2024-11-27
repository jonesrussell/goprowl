package crawlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/goprowl/metrics"
)

// ConfigOptions holds command-line parameters for the crawler
type ConfigOptions struct {
	URL      string
	MaxDepth int
	Debug    bool
}

// Config holds crawler configuration
type Config struct {
	ConfigOptions  // Embed the options
	AllowedDomains []string
	UserAgent      string
	Parallelism    int
	RequestDelay   time.Duration
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
func NewCrawlerFromConfig(config *Config, metrics *metrics.ComponentMetrics) *CollyCrawler {
	opts := []colly.CollectorOption{
		colly.MaxDepth(config.MaxDepth),
		colly.Async(true),
		colly.UserAgent(config.UserAgent),
	}

	if config.Debug {
		opts = append(opts, colly.Debugger(&debug.LogDebugger{}))
	}

	c := colly.NewCollector(opts...)

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

	return &CollyCrawler{
		collector: c,
		metrics:   metrics,
	}
}
