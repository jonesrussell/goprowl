package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "goprowl"

// Core metrics
var (
	ActiveRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "active_requests",
			Help:      "Number of active crawl requests",
		},
	)

	PagesProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "pages_processed_total",
			Help:      "Total number of pages processed",
		},
	)

	ErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "errors_total",
			Help:      "Total number of errors encountered",
		},
	)
)

// PushGatewayClient defines the interface for pushing metrics
type PushGatewayClient interface {
	Push(ctx context.Context) error
	StartRequest(crawlerID string) error
	EndRequest(crawlerID string) error
	RecordCrawlMetrics(ctx context.Context, crawlerID, url, status string, duration time.Duration, pagesProcessed int) error
}

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
