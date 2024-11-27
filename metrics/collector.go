package metrics

import (
	"context"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	dto "github.com/prometheus/client_model/go"
	"go.uber.org/zap"
)

var (
	collector   *MetricsCollector
	collectorMu sync.RWMutex
)

const (
	namespace = "goprowl"
	component = "component_id"
)

// MetricsCollector provides a central collection point for all application metrics
type MetricsCollector struct {
	mu          sync.RWMutex
	pushgateway string
	logger      *zap.Logger

	// Crawler metrics
	totalActiveRequests *prometheus.GaugeVec
	totalPagesProcessed *prometheus.CounterVec
	totalErrors         *prometheus.CounterVec
	responseSizes       *prometheus.HistogramVec
	requestDurations    *prometheus.HistogramVec

	// List command metrics
	listOperationDuration *prometheus.HistogramVec
	listOperationErrors   *prometheus.CounterVec
	indexedDocuments      *prometheus.GaugeVec

	// New detailed metrics
	crawlDuration *prometheus.HistogramVec
	pageDepth     *prometheus.HistogramVec
	contentTypes  *prometheus.CounterVec
	statusCodes   *prometheus.CounterVec
	linkCount     *prometheus.HistogramVec

	// Add other application metrics here
}

func NewMetricsCollector(config Config, logger *zap.Logger) (*MetricsCollector, error) {
	collectorMu.Lock()
	defer collectorMu.Unlock()

	if collector != nil {
		return collector, nil
	}

	logger.Info("initializing metrics collector")

	registry := prometheus.NewRegistry()
	collector = &MetricsCollector{
		pushgateway: config.PushgatewayURL,
		logger:      logger,
		totalPagesProcessed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "goprowl",
			Name:      "pages_processed_total",
			Help:      "Total number of pages processed",
		}, []string{"component_id", "component_type"}),
		totalErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "errors_total",
			Help:      "Total number of crawler errors",
		}, []string{component}),
		responseSizes: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "response_sizes_bytes",
			Help:      "Size of HTTP responses in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		}, []string{component}),
		requestDurations: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds",
			Buckets:   prometheus.DefBuckets,
		}, []string{component}),
		listOperationDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "list_operation_duration_seconds",
			Help:      "Duration of list operations in seconds",
			Buckets:   prometheus.DefBuckets,
		}, []string{component}),
		listOperationErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "list_operation_errors_total",
			Help:      "Total number of list operation errors",
		}, []string{component}),
		indexedDocuments: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "indexed_documents_total",
			Help:      "Total number of indexed documents",
		}, []string{component}),
		crawlDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "crawl_duration_seconds",
			Help:      "Duration of crawl operations",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 10),
		}, []string{component}),
		pageDepth: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "page_depth",
			Help:      "Depth of crawled pages",
			Buckets:   []float64{1, 2, 3, 4, 5, 10},
		}, []string{component}),
		contentTypes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "content_types_total",
			Help:      "Count of different content types encountered",
		}, []string{component, "content_type"}),
		statusCodes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "status_codes_total",
			Help:      "Count of HTTP status codes",
		}, []string{component, "code"}),
		linkCount: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "links_per_page",
			Help:      "Number of links found per page",
			Buckets:   prometheus.LinearBuckets(0, 10, 10),
		}, []string{component}),
		totalActiveRequests: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "active_requests",
			Help:      "Number of currently active requests",
		}, []string{"component_id", "component_type"}),
	}

	collector.totalPagesProcessed.WithLabelValues("crawler", "crawler").Add(0)
	collector.totalActiveRequests.WithLabelValues("crawler", "crawler").Set(0)

	if err := registerMetrics(collector, registry, logger); err != nil {
		return nil, fmt.Errorf("failed to register metrics: %w", err)
	}

	prometheus.DefaultRegisterer = registry
	prometheus.DefaultGatherer = registry

	logger.Info("metrics collector initialized successfully")
	return collector, nil
}

func registerMetrics(c *MetricsCollector, registry *prometheus.Registry, logger *zap.Logger) error {
	metrics := []prometheus.Collector{
		c.totalActiveRequests,
		c.totalPagesProcessed,
		c.totalErrors,
		c.responseSizes,
		c.requestDurations,
		c.listOperationDuration,
		c.listOperationErrors,
		c.indexedDocuments,
		c.crawlDuration,
		c.pageDepth,
		c.contentTypes,
		c.statusCodes,
		c.linkCount,
	}

	for _, metric := range metrics {
		if err := registry.Register(metric); err != nil {
			logger.Error("failed to register metric", zap.Error(err))
			return err
		}
		logger.Info("registered metric successfully", zap.String("metric", fmt.Sprintf("%T", metric)))
	}

	return nil
}

// PushMetrics pushes all metrics to Pushgateway with context
func (c *MetricsCollector) PushMetrics(ctx context.Context, job string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	done := make(chan error, 1)

	go func() {
		pusher := push.New(c.pushgateway, job)
		done <- pusher.Push()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Add these helper methods to MetricsCollector
func (c *MetricsCollector) IncrementPagesProcessedWithLabel(componentID string) {
	c.totalPagesProcessed.WithLabelValues(componentID).Inc()
}

func (c *MetricsCollector) GetPagesProcessed(componentID string) float64 {
	m := &dto.Metric{}
	c.totalPagesProcessed.WithLabelValues(componentID).Write(m)
	return *m.Counter.Value
}

// GetComponentMetrics creates a new ComponentMetrics instance
func (c *MetricsCollector) GetComponentMetrics(componentType string) *ComponentMetrics {
	return NewComponentMetrics(c, componentType)
}
