package metrics

import (
	"context"
	"time"
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
