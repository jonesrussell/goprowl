package metrics

import (
	"context"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	registrationOnce sync.Once
	collector        *MetricsCollector
	collectorMu      sync.RWMutex
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
	collectorMu.Lock()
	defer collectorMu.Unlock()

	// Return existing collector if already initialized
	if collector != nil {
		return collector, nil
	}

	collector = &MetricsCollector{
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

	// Register metrics with prometheus only once
	var registerErr error
	registrationOnce.Do(func() {
		// Use MustRegister inside a recover to convert panic to error
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					registerErr = err
				} else {
					registerErr = fmt.Errorf("metrics registration failed: %v", r)
				}
			}
		}()

		prometheus.MustRegister(
			collector.totalActiveRequests,
			collector.totalPagesProcessed,
			collector.totalErrors,
			collector.responseSizes,
			collector.requestDurations,
		)
	})

	if registerErr != nil {
		return nil, registerErr
	}

	return collector, nil
}

// PushMetrics pushes all metrics to Pushgateway with context
func (c *MetricsCollector) PushMetrics(ctx context.Context, job string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create a channel to handle timeout
	done := make(chan error, 1)

	go func() {
		pusher := push.New(c.pushgateway, job)
		done <- pusher.Push()
	}()

	// Wait for either the push to complete or context to timeout
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
