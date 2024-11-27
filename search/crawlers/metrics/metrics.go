package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// CrawlerMetrics holds all crawler-related metrics
type CrawlerMetrics struct {
	RequestDuration prometheus.Histogram
	ResponseSize    prometheus.Histogram
	ErrorCount      prometheus.Counter
	PagesProcessed  prometheus.Counter
	ActiveRequests  prometheus.Gauge
}

// NewCrawlerMetrics creates and registers crawler metrics
func NewCrawlerMetrics() *CrawlerMetrics {
	return &CrawlerMetrics{
		RequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "crawler_request_duration_seconds",
			Help:    "Time spent processing requests",
			Buckets: prometheus.DefBuckets,
		}),
		ResponseSize: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "crawler_response_size_bytes",
			Help:    "Size of HTTP responses",
			Buckets: []float64{1024, 10 * 1024, 100 * 1024, 1024 * 1024},
		}),
		ErrorCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "crawler_errors_total",
			Help: "Total number of crawler errors",
		}),
		PagesProcessed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "crawler_pages_processed_total",
			Help: "Total number of pages processed",
		}),
		ActiveRequests: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "crawler_active_requests",
			Help: "Number of currently active requests",
		}),
	}
}
