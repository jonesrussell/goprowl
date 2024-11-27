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

// Constants for label keys
const (
	LabelComponentID   = "component_id"
	LabelComponentType = "component_type"
	LabelCrawlerID     = "crawler_id"
	LabelStatus        = "status"
	LabelURL           = "url"
)

// CrawlerMetrics contains Pushgateway-specific metrics
var (
	activeRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "goprowl",
			Name:      "active_requests",
			Help:      "Number of currently active crawler requests",
		},
		[]string{LabelComponentID, LabelComponentType},
	)

	crawlerCompletionTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "goprowl",
			Subsystem: "crawler",
			Name:      "completion_timestamp",
			Help:      "Timestamp when a crawler completed its task",
		},
		[]string{LabelCrawlerID, LabelStatus, LabelURL},
	)

	crawlerPagesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "goprowl",
			Subsystem: "crawler",
			Name:      "pages_processed_total",
			Help:      "Total number of pages processed by crawler",
		},
		[]string{LabelCrawlerID},
	)

	crawlerDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "goprowl",
			Subsystem: "crawler",
			Name:      "duration_seconds",
			Help:      "Duration of crawler run in seconds",
		},
		[]string{LabelCrawlerID, LabelURL},
	)
)

// pushGatewayClientImpl implements the PushGatewayClient interface
type pushGatewayClientImpl struct {
	logger   *zap.Logger
	pusher   *push.Pusher
	registry *prometheus.Registry
}

// NewPushGatewayClient creates a new PushGatewayClient
func NewPushGatewayClient(lc fx.Lifecycle, logger *zap.Logger, config Config) (PushGatewayClient, error) {
	registry := prometheus.NewRegistry()

	// Register metrics with the new registry
	metrics := []prometheus.Collector{
		activeRequests,
		crawlerCompletionTime,
		crawlerPagesProcessed,
		crawlerDuration,
	}

	for _, m := range metrics {
		if err := registry.Register(m); err != nil {
			return nil, fmt.Errorf("failed to register metric: %w", err)
		}
	}

	pusher := push.New(config.PushgatewayURL, "goprowl").
		Gatherer(registry)

	client := &pushGatewayClientImpl{
		logger:   logger,
		pusher:   pusher,
		registry: registry,
	}

	// Add lifecycle hooks
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return client.Push(ctx)
		},
	})

	return client, nil
}

// Push implements PushGatewayClient
func (c *pushGatewayClientImpl) Push(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if err := c.pusher.Push(); err != nil {
			c.logger.Debug("failed to push metrics to Pushgateway (optional)",
				zap.Error(err))
			return nil
		}
		c.logger.Debug("successfully pushed metrics to Pushgateway")
		return nil
	}
}

// StartRequest implements PushGatewayClient
func (c *pushGatewayClientImpl) StartRequest(crawlerID string) error {
	activeRequests.WithLabelValues(crawlerID, "crawler").Inc()
	return c.Push(context.Background())
}

// EndRequest implements PushGatewayClient
func (c *pushGatewayClientImpl) EndRequest(crawlerID string) error {
	activeRequests.WithLabelValues(crawlerID, "crawler").Dec()
	return c.Push(context.Background())
}

// RecordCrawlMetrics implements PushGatewayClient
func (c *pushGatewayClientImpl) RecordCrawlMetrics(ctx context.Context, crawlerID, url, status string, duration time.Duration, pagesProcessed int) error {
	// Set completion timestamp
	crawlerCompletionTime.WithLabelValues(crawlerID, status, url).Set(float64(time.Now().Unix()))

	// Set pages processed
	crawlerPagesProcessed.WithLabelValues(crawlerID).Add(float64(pagesProcessed))

	// Set duration
	crawlerDuration.WithLabelValues(crawlerID, url).Set(duration.Seconds())

	// Push metrics
	if err := c.Push(ctx); err != nil {
		return fmt.Errorf("failed to record crawl metrics: %w", err)
	}

	c.logger.Info("recorded crawl metrics",
		zap.String("crawler_id", crawlerID),
		zap.String("url", url),
		zap.String("status", status),
		zap.Duration("duration", duration),
		zap.Int("pages_processed", pagesProcessed),
	)

	return nil
}
