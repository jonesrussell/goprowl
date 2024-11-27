package metrics

import (
	"time"
)

// Config holds the configuration for the metrics collector
type Config struct {
	PushgatewayURL string
	PushInterval   time.Duration
	MetricsPort    string
}
