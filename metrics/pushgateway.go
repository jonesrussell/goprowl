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
	logger *zap.Logger
	pusher *push.Pusher
}

type PushGatewayParams struct {
	fx.In
	Logger *zap.Logger
}

func NewPushGatewayClient(p PushGatewayParams) (*PushGatewayClient, error) {
	// Register metrics
	prometheus.MustRegister(crawlerCompletionTime)
	prometheus.MustRegister(crawlerPagesProcessed)
	prometheus.MustRegister(crawlerDuration)

	pusher := push.New("pushgateway:9091", "goprowl").
		Gatherer(prometheus.DefaultGatherer)

	return &PushGatewayClient{
		logger: p.Logger,
		pusher: pusher,
	}, nil
}

// RecordCrawlCompletion records the completion of a crawl operation
func (c *PushGatewayClient) RecordCrawlCompletion(ctx context.Context, crawlerID, status, url string, duration time.Duration, pagesProcessed int) error {
	// Record metrics
	crawlerCompletionTime.WithLabelValues(crawlerID, status, url).Set(float64(time.Now().Unix()))
	crawlerPagesProcessed.WithLabelValues(crawlerID).Add(float64(pagesProcessed))
	crawlerDuration.WithLabelValues(crawlerID, url).Set(duration.Seconds())

	// Push to Pushgateway
	if err := c.pusher.Add(); err != nil {
		return fmt.Errorf("failed to push metrics: %w", err)
	}

	c.logger.Info("pushed crawler metrics to Pushgateway",
		zap.String("crawler_id", crawlerID),
		zap.String("status", status),
		zap.String("url", url),
		zap.Duration("duration", duration),
		zap.Int("pages_processed", pagesProcessed),
	)

	return nil
}
