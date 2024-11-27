package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("metrics",
	fx.Provide(
		NewRegistry,
		NewMetricsCollector,
		NewPushGatewayClient,
		func(collector *MetricsCollector, logger *zap.Logger) *ComponentMetrics {
			return NewComponentMetrics(collector, "crawler", logger)
		},
		NewConfig,
	),
)

// NewConfig provides the metrics configuration
func NewConfig() Config {
	return Config{
		PushgatewayURL: "http://localhost:9091",
		PushInterval:   5 * time.Second,
		MetricsPort:    ":8080",
	}
}

func NewRegistry() (*prometheus.Registry, error) {
	registry := prometheus.NewRegistry()

	// Register default collectors
	registry.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	)

	return registry, nil
}
