package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
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

// Modify the registry provider to register metrics
func NewRegistry() *prometheus.Registry {
	// Create a new registry
	registry := prometheus.NewRegistry()
	if registry == nil {
		panic("failed to create prometheus registry")
	}

	// Only register the Go and Process collectors
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	return registry
}

// Module initializes the metrics collection module
var Module = fx.Module("metrics",
	fx.Provide(
		NewDefaultConfig,
		NewRegistry,
		fx.Annotate(
			NewMetricsCollector,
			fx.ParamTags(``, ``, `name:"metrics_registry"`),
		),
		fx.Annotate(
			NewComponentMetrics,
			fx.As(new(ComponentMetricsProvider)),
		),
	),
)
