package crawlers

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/goprowl/metrics"
	"go.uber.org/zap"
)

type CollyCrawler struct {
	collector *colly.Collector
	metrics   *metrics.ComponentMetrics
	id        string
	logger    *zap.Logger
	cfg       *Config
}

func NewCollyCrawler(
	logger *zap.Logger,
	collector *colly.Collector,
	metrics *metrics.ComponentMetrics,
	cfg *Config,
) (Crawler, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	uniqueID := fmt.Sprintf("crawler-%d", time.Now().UnixNano())

	crawler := &CollyCrawler{
		collector: collector,
		metrics:   metrics,
		id:        uniqueID,
		logger:    logger,
		cfg:       cfg,
	}

	setupCallbacks(collector, metrics, logger)

	return crawler, nil
}

func setupCallbacks(c *colly.Collector, m *metrics.ComponentMetrics, logger *zap.Logger) {
	c.OnRequest(func(r *colly.Request) {
		m.IncrementActiveRequests()
		logger.Info("starting request",
			zap.String("url", r.URL.String()),
		)
	})

	c.OnError(func(r *colly.Response, err error) {
		m.DecrementActiveRequests()
		m.IncrementErrorCount()
		logger.Error("error visiting url",
			zap.String("url", r.Request.URL.String()),
			zap.Error(err),
		)
	})

	c.OnResponse(func(r *colly.Response) {
		m.DecrementActiveRequests()
		m.IncrementPagesProcessed()
		m.ObserveResponseSize(float64(len(r.Body)))
		logger.Info("received response",
			zap.String("url", r.Request.URL.String()),
			zap.Int("size", len(r.Body)),
			zap.Int("status", r.StatusCode),
		)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absLink := e.Request.AbsoluteURL(link)
		if absLink != "" {
			logger.Debug("found link",
				zap.String("source_url", e.Request.URL.String()),
				zap.String("link", absLink),
			)
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
	c.logger.Info("starting crawl",
		zap.String("crawler_id", c.id),
		zap.String("url", startURL),
		zap.Int("depth", depth),
	)

	// Verify URL is valid
	parsedURL, err := url.Parse(startURL)
	if err != nil {
		c.logger.Error("invalid url",
			zap.String("url", startURL),
			zap.Error(err),
		)
		return fmt.Errorf("invalid URL %s: %w", startURL, err)
	}

	// Allow the domain we're crawling
	c.collector.AllowedDomains = []string{parsedURL.Host}
	c.logger.Info("set allowed domain",
		zap.String("domain", parsedURL.Host),
	)

	// Configure parallel requests using config values
	c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: c.cfg.Parallelism,
		RandomDelay: c.cfg.RequestDelay,
	})

	c.logger.Info("initiating crawl")
	if err = c.collector.Visit(startURL); err != nil {
		c.logger.Error("failed to start crawl",
			zap.String("url", startURL),
			zap.Error(err),
		)
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
		c.logger.Info("crawl completed successfully",
			zap.String("url", startURL),
			zap.Int("depth", depth),
		)
	case <-time.After(2 * time.Minute):
		c.logger.Error("crawl timed out",
			zap.String("url", startURL),
			zap.Duration("timeout", 2*time.Minute),
		)
		return fmt.Errorf("crawl timed out after 2 minutes")
	case <-ctx.Done():
		c.logger.Warn("crawl cancelled",
			zap.String("url", startURL),
			zap.Error(ctx.Err()),
		)
		return ctx.Err()
	}

	return nil
}
