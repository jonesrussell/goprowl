package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	ActiveRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "active_requests",
			Help:      "Number of active crawl requests",
		},
		[]string{"component_type"},
	)

	PagesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "pages_processed_total",
			Help:      "Total number of pages processed",
		},
		[]string{"component_type"},
	)

	ErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "errors_total",
			Help:      "Total number of errors encountered",
		},
		[]string{"component_type"},
	)
)

// QueryResponse represents the structure of a Prometheus query response
type QueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// PushGatewayClient defines the interface for interacting with a Pushgateway
type PushGatewayClient interface {
	// StartRequest starts tracking a new request by its ID
	StartRequest(id string) error

	// EndRequest ends tracking for the specified request ID
	EndRequest(id string) error

	// RecordCrawlMetrics records metrics for a crawl operation
	RecordCrawlMetrics(ctx context.Context, id, url, status string, duration time.Duration, pagesVisited int) error

	// Push pushes the collected metrics to the Pushgateway
	Push(ctx context.Context) error
}

// ComponentMetricsProvider defines the interface for providing component-level metrics
type ComponentMetricsProvider interface {
	// Counter returns or creates a new counter metric
	Counter(name, help string, labelNames ...string) *prometheus.Counter

	// Gauge returns or creates a new gauge metric
	Gauge(name, help string, labelNames ...string) *prometheus.Gauge

	// Histogram returns or creates a new histogram metric
	Histogram(name, help string, buckets []float64, labelNames ...string) *prometheus.Histogram
}
