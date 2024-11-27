package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

// Provide default config if none is supplied
func NewDefaultConfig() Config {
	return Config{
		PushgatewayURL: "http://pushgateway:9091",
		PushInterval:   15 * time.Second,
		MetricsPort:    ":8085",
	}
}

var Module = fx.Module("metrics",
	fx.Provide(
		fx.Annotate(
			NewDefaultConfig,
			fx.ResultTags(`optional:"true"`), // Mark as optional to allow override
		),
		NewMetricsCollector,
		func(collector *MetricsCollector) *ComponentMetrics {
			return NewComponentMetrics(collector, "goprowl")
		},
		NewPushGatewayClient,
		func() *prometheus.Registry {
			return prometheus.NewRegistry()
		},
		NewMetricsServer,
	),
)
