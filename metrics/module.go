package metrics

import (
	"context"

	"go.uber.org/fx"
)

var Module = fx.Module("metrics",
	fx.Provide(
		NewMetricsCollector,
		func(collector *MetricsCollector) *ComponentMetrics {
			return NewComponentMetrics(collector, "crawler")
		},
	),
	fx.Invoke(registerMetricsHandlers),
)

func registerMetricsHandlers(lc fx.Lifecycle, collector *MetricsCollector) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return collector.PushMetrics("goprowl")
		},
	})
}
