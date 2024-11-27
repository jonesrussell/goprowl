package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// CrawlerMetrics contains Pushgateway-specific metrics
var (
	activeRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "goprowl",
			Name:      "active_requests",
			Help:      "Number of currently active crawler requests",
		},
		[]string{"crawler_id"},
	)

	crawlerCompletionTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "goprowl",
			Subsystem: "crawler",
			Name:      "completion_timestamp",
			Help:      "Timestamp when a crawler completed its task",
		},
		[]string{"crawler_id", "status", "url"},
	)

	crawlerPagesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "goprowl",
			Subsystem: "crawler",
			Name:      "pages_processed_total",
			Help:      "Total number of pages processed by crawler",
		},
		[]string{"crawler_id"},
	)

	crawlerDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "goprowl",
			Subsystem: "crawler",
			Name:      "duration_seconds",
			Help:      "Duration of crawler run in seconds",
		},
		[]string{"crawler_id", "url"},
	)
)

type PushGatewayClient struct {
	logger   *zap.Logger
	pusher   *push.Pusher
	registry *prometheus.Registry
}

func NewPushGatewayClient(lc fx.Lifecycle, logger *zap.Logger) (*PushGatewayClient, error) {
	// Create a new registry instead of using the default one
	registry := prometheus.NewRegistry()

	// Register metrics with the new registry
	activeRequests := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "goprowl_active_requests",
			Help: "Number of currently active crawler requests",
		},
		[]string{"crawler_id"},
	)

	if err := registry.Register(activeRequests); err != nil {
		return nil, fmt.Errorf("failed to register active_requests metric: %w", err)
	}

	// Create pusher with job name "goprowl"
	pusher := push.New("pushgateway:9091", "goprowl").
		Gatherer(prometheus.DefaultGatherer)

	client := &PushGatewayClient{
		logger:   logger,
		pusher:   pusher,
		registry: registry,
	}

	// Add lifecycle hooks
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return client.pusher.Push()
		},
	})

	return client, nil
}

// RecordCrawlMetrics records metrics for a crawl operation
func (c *PushGatewayClient) RecordCrawlMetrics(ctx context.Context, crawlerID string, url string, status string, duration time.Duration, pagesProcessed int) error {
	// Set completion timestamp
	crawlerCompletionTime.WithLabelValues(crawlerID, status, url).Set(float64(time.Now().Unix()))

	// Set pages processed
	crawlerPagesProcessed.WithLabelValues(crawlerID).Add(float64(pagesProcessed))

	// Set duration
	crawlerDuration.WithLabelValues(crawlerID, url).Set(duration.Seconds())

	// Push metrics to Pushgateway
	if err := c.pusher.Push(); err != nil {
		return fmt.Errorf("failed to push metrics: %w", err)
	}

	c.logger.Info("pushed crawler metrics",
		zap.String("crawler_id", crawlerID),
		zap.String("url", url),
		zap.String("status", status),
		zap.Duration("duration", duration),
		zap.Int("pages_processed", pagesProcessed),
	)

	return nil
}

// StartRequest increments the active request counter
func (c *PushGatewayClient) StartRequest(crawlerID string) error {
	activeRequests.WithLabelValues(crawlerID).Inc()
	if err := c.pusher.Push(); err != nil {
		return fmt.Errorf("failed to push start request metric: %w", err)
	}
	return nil
}

// EndRequest decrements the active request counter
func (c *PushGatewayClient) EndRequest(crawlerID string) error {
	activeRequests.WithLabelValues(crawlerID).Dec()
	if err := c.pusher.Push(); err != nil {
		return fmt.Errorf("failed to push end request metric: %w", err)
	}
	return nil
}
