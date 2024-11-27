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

// CollyCrawler is a struct that encapsulates the necessary components for web crawling
type CollyCrawler struct {
	collector    *colly.Collector
	metrics      *metrics.ComponentMetrics
	pushgateway  metrics.PushGatewayClient
	id           string
	logger       *zap.Logger
	cfg          *Config
	startTime    time.Time
	pagesVisited int
}

// NewCollyCrawler creates a new instance of CollyCrawler
func NewCollyCrawler(
	logger *zap.Logger,
	collector *colly.Collector,
	metrics *metrics.ComponentMetrics,
	pushgateway metrics.PushGatewayClient,
	cfg *Config,
) (Crawler, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	uniqueID := fmt.Sprintf("component-%d", time.Now().UnixNano())

	crawler := &CollyCrawler{
		collector:    collector,
		metrics:      metrics,
		pushgateway:  pushgateway,
		id:           uniqueID,
		logger:       logger,
		cfg:          cfg,
		pagesVisited: 0,
	}

	setupCallbacks(collector, metrics, logger)

	return crawler, nil
}

// setupCallbacks sets up the necessary callbacks for the Colly collector
func setupCallbacks(c *colly.Collector, m *metrics.ComponentMetrics, logger *zap.Logger) {
	c.OnRequest(func(r *colly.Request) {
		m.IncrementActiveRequests()
		m.IncrementActiveRequestsWithLabel("crawler")
		logger.Info("starting request", zap.String("url", r.URL.String()))
	})

	c.OnError(func(r *colly.Response, err error) {
		m.DecrementActiveRequests()
		m.IncrementErrorCountWithLabel("crawler")
		logger.Error("error visiting url", zap.String("url", r.Request.URL.String()), zap.Error(err))
	})

	c.OnResponse(func(r *colly.Response) {
		m.DecrementActiveRequests()
		m.IncrementPagesProcessedWithLabel("crawler")
		m.ObserveResponseSize(float64(len(r.Body)))
		logger.Info("received response", zap.String("url", r.Request.URL.String()), zap.Int("size", len(r.Body)), zap.Int("status", r.StatusCode))
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absLink := e.Request.AbsoluteURL(link)
		if absLink != "" {
			logger.Debug("found link", zap.String("source_url", e.Request.URL.String()), zap.String("link", absLink))
			e.Request.Visit(absLink)
		}
	})
}

// GetID returns the crawler's unique identifier
func (c *CollyCrawler) GetID() string {
	return c.id
}

// Crawl implements the Crawler interface and starts the crawling process
func (c *CollyCrawler) Crawl(ctx context.Context, startURL string, depth int) error {
	// Create a default handler that just logs
	defaultHandler := func(ctx context.Context, result *CrawlResult) error {
		c.logger.Info("crawled page", zap.String("url", result.URL), zap.String("title", result.Title))
		return nil
	}

	return c.CrawlWithHandler(ctx, startURL, depth, defaultHandler)
}

// CrawlStats contains statistics for the crawl process
type CrawlStats struct {
	PagesVisited     int
	ErrorCount       int
	CurrentDepth     int
	StartTime        time.Time
	LastUpdateTime   time.Time
	ContentTypeStats map[string]int
	StatusCodeStats  map[int]int
}

// CrawlWithHandler implements the Crawler interface and handles the crawling process with a custom handler
func (c *CollyCrawler) CrawlWithHandler(ctx context.Context, startURL string, depth int, handler PageHandler) error {
	// Initialize metrics at start
	c.metrics.ResetActiveRequests()
	c.startTime = time.Now()

	stats := &CrawlStats{
		StartTime:        time.Now(),
		ContentTypeStats: make(map[string]int),
		StatusCodeStats:  make(map[int]int),
	}

	// Add structured logging
	statusLogger := c.logger.With(zap.String("crawler_id", c.id), zap.String("url", startURL), zap.Int("depth", depth))

	// Log initial state
	statusLogger.Info("crawl started", zap.Time("start_time", stats.StartTime))

	// Add periodic status updates
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				stats.LastUpdateTime = time.Now()
				statusLogger.Info("crawl status",
					zap.Int("pages_visited", stats.PagesVisited),
					zap.Int("errors", stats.ErrorCount),
					zap.Int("current_depth", stats.CurrentDepth),
					zap.Duration("elapsed_time", time.Since(stats.StartTime)),
					zap.Any("content_types", stats.ContentTypeStats),
					zap.Any("status_codes", stats.StatusCodeStats),
				)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Start request tracking
	if err := c.pushgateway.StartRequest(c.id); err != nil {
		c.logger.Error("failed to track request start", zap.Error(err))
	}
	defer func() {
		if err := c.pushgateway.EndRequest(c.id); err != nil {
			c.logger.Error("failed to track request end", zap.Error(err))
		}
	}()

	// Verify URL is valid
	parsedURL, err := url.Parse(startURL)
	if err != nil {
		c.logger.Error("invalid url", zap.String("url", startURL), zap.Error(err))
		return fmt.Errorf("invalid URL %s: %w", startURL, err)
	}

	// Allow the domain we're crawling
	c.collector.AllowedDomains = []string{parsedURL.Host}

	// Configure collector callbacks for handling pages
	c.collector.OnHTML("html", func(e *colly.HTMLElement) {
		c.pagesVisited++
		result := &CrawlResult{
			URL:       e.Request.URL.String(),
			Title:     e.ChildText("title"),
			Content:   e.Text,
			Links:     e.ChildAttrs("a[href]", "href"),
			CreatedAt: time.Now().Format(time.RFC3339),
		}

		c.logger.Debug("processing page", zap.String("url", result.URL), zap.String("title", result.Title), zap.Int("links_count", len(result.Links)))

		if err := handler(ctx, result); err != nil {
			c.logger.Error("handler failed", zap.String("url", result.URL), zap.Error(err))
		}
	})

	// Configure parallel requests using config values
	c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: c.cfg.Parallelism,
		RandomDelay: c.cfg.RequestDelay,
	})

	if err = c.collector.Visit(startURL); err != nil {
		c.logger.Error("failed to start crawl", zap.String("url", startURL), zap.Error(err))
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
		duration := time.Since(c.startTime)
		err := c.pushgateway.RecordCrawlMetrics(
			ctx,
			c.id,
			startURL,
			"completed",
			duration,
			c.pagesVisited,
		)
		if err != nil {
			c.logger.Error("failed to push metrics", zap.Error(err))
		}

		c.logger.Info("crawl completed successfully", zap.String("url", startURL), zap.Int("depth", depth), zap.Duration("duration", duration), zap.Int("pages_visited", c.pagesVisited))
	case <-time.After(2 * time.Minute):
		c.logger.Error("crawl timed out", zap.String("url", startURL), zap.Duration("timeout", 2*time.Minute))
		return fmt.Errorf("crawl timed out after 2 minutes")
	case <-ctx.Done():
		c.logger.Warn("crawl cancelled", zap.String("url", startURL), zap.Error(ctx.Err()))
		return ctx.Err()
	}

	// Update metrics periodically
	statusTicker := time.NewTicker(5 * time.Second)
	defer statusTicker.Stop()

	go func() {
		for {
			select {
			case <-statusTicker.C:
				if err := c.pushgateway.Push(ctx); err != nil {
					c.logger.Error("failed to push metrics", zap.Error(err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
