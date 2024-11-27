package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

// NewDefaultConfig provides default config if none is supplied
func NewDefaultConfig() Config {
	return Config{
		PushgatewayURL: "http://pushgateway:9091",
		PushInterval:   15 * time.Second,
		MetricsPort:    ":8085",
	}
}

// Constants for annotations and tags
const (
	OptionalTrueTag = `optional:"true"`
)

// Module initializes the metrics collection module
var Module = fx.Module("metrics",
	fx.Provide(
		// Annotate the NewDefaultConfig function to make it optional for override
		fx.Annotate(
			NewDefaultConfig,
			fx.ResultTags(OptionalTrueTag), // Mark as optional to allow override
		),
		NewMetricsCollector,
		// Provide a new ComponentMetrics instance
		func(collector *MetricsCollector) *ComponentMetrics {
			return NewComponentMetrics(collector, "goprowl")
		},
		NewPushGatewayClient,
		// Provide a new prometheus Registry
		func() *prometheus.Registry {
			return prometheus.NewRegistry()
		},
		NewMetricsServer,
	),
)
