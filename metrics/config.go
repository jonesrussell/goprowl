package metrics

import (
	"time"
)

type Config struct {
	PushgatewayURL string
	PushInterval   time.Duration
	// Add other metrics configuration as needed
}
