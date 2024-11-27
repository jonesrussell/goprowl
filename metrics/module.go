package metrics

import (
	"context"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
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
	fx.Invoke(registerMetricsHandlers),
)

func registerMetricsHandlers(lc fx.Lifecycle, collector *MetricsCollector, logger *zap.Logger) {
	// Push metrics periodically
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting metrics pusher",
				zap.String("pushgateway", collector.pushgateway))

			// Initial push
			if err := collector.PushMetrics(ctx, "goprowl"); err != nil {
				logger.Error("failed to push initial metrics", zap.Error(err))
				// Don't fail startup for metrics
				return nil
			}

			// Start periodic push
			go func() {
				ticker := time.NewTicker(15 * time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						if err := collector.PushMetrics(ctx, "goprowl"); err != nil {
							logger.Error("failed to push metrics", zap.Error(err))
						}
					}
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Final push before shutdown
			if err := collector.PushMetrics(ctx, "goprowl"); err != nil {
				logger.Error("failed to push final metrics", zap.Error(err))
			}
			return nil
		},
	})
}
