package metrics

import (
	"context"
	"time"

	"go.uber.org/fx"
)

// Provide default config if none is supplied
func NewDefaultConfig() Config {
	return Config{
		PushgatewayURL: "http://localhost:9091",
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
	),
	fx.Invoke(registerMetricsHandlers),
)

func registerMetricsHandlers(lc fx.Lifecycle, collector *MetricsCollector) {
	lc.Append(fx.Hook{
		OnStop: func(baseCtx context.Context) error {
			ctx, cancel := context.WithTimeout(baseCtx, 5*time.Second)
			defer cancel()

			if err := collector.PushMetrics(ctx, "goprowl"); err != nil {
				return nil
			}
			return nil
		},
	})
}
