package metrics

import (
	"time"

	"go.uber.org/fx"
)

// Provide default config if none is supplied
func NewDefaultConfig() Config {
	return Config{
		PushgatewayURL: "http://pushgateway:9091",
		PushInterval:   15 * time.Second,
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
	),
)
