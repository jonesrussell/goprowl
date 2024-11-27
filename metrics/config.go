package metrics

import (
	"time"
)

type Config struct {
	PushgatewayURL string
	PushInterval   time.Duration
	MetricsPort    string
	// Add other metrics configuration as needed
}
