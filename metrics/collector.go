package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

// MetricsCollector provides a central collection point for all application metrics
type MetricsCollector struct {
	mu          sync.RWMutex
	pushgateway string

	// Crawler metrics
	totalActiveRequests *prometheus.GaugeVec
	totalPagesProcessed *prometheus.CounterVec
	totalErrors         *prometheus.CounterVec
	responseSizes       *prometheus.HistogramVec
	requestDurations    *prometheus.HistogramVec

	// Add other application metrics here
}

func NewMetricsCollector(config Config) (*MetricsCollector, error) {
	collector := &MetricsCollector{
		pushgateway: config.PushgatewayURL,
		totalActiveRequests: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "goprowl_active_requests",
			Help: "Number of currently active crawler requests",
		}, []string{"component_id"}),
		totalPagesProcessed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "goprowl_pages_processed_total",
			Help: "Total number of pages processed",
		}, []string{"component_id"}),
		totalErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "goprowl_errors_total",
			Help: "Total number of crawler errors",
		}, []string{"component_id"}),
		responseSizes: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "goprowl_response_sizes_bytes",
			Help:    "Size of HTTP responses in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		}, []string{"component_id"}),
		requestDurations: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "goprowl_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"component_id"}),
	}

	// Register metrics with prometheus
	prometheus.MustRegister(
		collector.totalActiveRequests,
		collector.totalPagesProcessed,
		collector.totalErrors,
		collector.responseSizes,
		collector.requestDurations,
	)

	return collector, nil
}

// PushMetrics pushes all metrics to Pushgateway
func (c *MetricsCollector) PushMetrics(job string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	pusher := push.New(c.pushgateway, job)
	return pusher.Push()
}
