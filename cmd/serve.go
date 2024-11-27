package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jonesrussell/goprowl/internal/app"
	"github.com/jonesrussell/goprowl/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewServeCmd() *cobra.Command {
	var (
		port  int
		debug bool
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the metrics dashboard",
		Long:  `Start the HTTP server to serve the metrics dashboard. Metrics are pushed to Pushgateway.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			stop := make(chan os.Signal, 1)
			signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

			app := fx.New(
				fx.Provide(func() (*zap.Logger, error) {
					config := zap.NewDevelopmentConfig()
					if debug {
						config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
					}
					return config.Build()
				}),
				app.Module,
				metrics.Module,
				fx.Invoke(func(lifecycle fx.Lifecycle, logger *zap.Logger, registry *prometheus.Registry) {
					logger.Debug("initialized metrics registry")

					mux := http.NewServeMux()

					// Register all handlers
					metrics.RegisterDashboard(mux)
					mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
					mux.HandleFunc("/api/v1/query", func(w http.ResponseWriter, r *http.Request) {
						query := r.URL.Query().Get("query")
						if query == "" {
							http.Error(w, "query parameter is required", http.StatusBadRequest)
							return
						}

						registryMetrics, err := registry.Gather()
						if err != nil {
							logger.Error("Failed to gather metrics", zap.Error(err))
							http.Error(w, "internal server error", http.StatusInternalServerError)
							return
						}

						var result []struct {
							Metric map[string]string `json:"metric"`
							Value  []interface{}     `json:"value"`
						}

						timestamp := float64(time.Now().Unix())

						for _, mf := range registryMetrics {
							metricName := *mf.Name
							if matchesQuery(metricName, query) {
								for _, m := range mf.Metric {
									var value float64
									switch {
									case m.Gauge != nil:
										value = *m.Gauge.Value
									case m.Counter != nil:
										value = *m.Counter.Value
									default:
										continue
									}

									result = append(result, struct {
										Metric map[string]string `json:"metric"`
										Value  []interface{}     `json:"value"`
									}{
										Metric: map[string]string{"__name__": metricName},
										Value: []interface{}{
											timestamp,
											fmt.Sprintf("%v", value),
										},
									})
								}
							}
						}

						if debug {
							logger.Debug("processed metric query",
								zap.String("query", query),
								zap.Int("total_metrics", len(registryMetrics)),
								zap.Int("matching_results", len(result)))
						}

						response := metrics.QueryResponse{
							Status: "success",
							Data: struct {
								ResultType string `json:"resultType"`
								Result     []struct {
									Metric map[string]string `json:"metric"`
									Value  []interface{}     `json:"value"`
								} `json:"result"`
							}{
								ResultType: "vector",
								Result:     result,
							},
						}

						w.Header().Set("Content-Type", "application/json")
						if err := json.NewEncoder(w).Encode(response); err != nil {
							logger.Error("Failed to encode response", zap.Error(err))
							http.Error(w, "failed to encode response", http.StatusInternalServerError)
							return
						}
					})

					mux.HandleFunc("/debug/metrics", func(w http.ResponseWriter, r *http.Request) {
						logger.Info("debug metrics requested")

						metrics, err := prometheus.DefaultGatherer.Gather()
						if err != nil {
							logger.Error("failed to gather metrics", zap.Error(err))
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}

						logger.Info("gathered metrics", zap.Int("count", len(metrics)))

						response := make(map[string]interface{})
						for _, mf := range metrics {
							name := *mf.Name
							logger.Debug("processing metric family",
								zap.String("name", name),
								zap.Int("metrics_count", len(mf.Metric)))

							if strings.HasPrefix(name, "goprowl_") {
								values := make([]map[string]interface{}, 0)
								for _, m := range mf.Metric {
									metric := make(map[string]interface{})

									// Add labels
									labels := make(map[string]string)
									for _, l := range m.Label {
										labels[*l.Name] = *l.Value
									}
									metric["labels"] = labels

									// Add value based on metric type
									if m.Gauge != nil {
										metric["value"] = *m.Gauge.Value
										metric["type"] = "gauge"
									} else if m.Counter != nil {
										metric["value"] = *m.Counter.Value
										metric["type"] = "counter"
									}

									values = append(values, metric)
									logger.Debug("added metric",
										zap.String("name", name),
										zap.Any("labels", labels),
										zap.Any("metric", metric))
								}
								response[name] = values
							}
						}

						w.Header().Set("Content-Type", "application/json")
						if err := json.NewEncoder(w).Encode(response); err != nil {
							logger.Error("failed to encode response", zap.Error(err))
							http.Error(w, "failed to encode response", http.StatusInternalServerError)
							return
						}

						logger.Info("debug metrics returned successfully",
							zap.Int("metrics_count", len(response)))
					})

					server := &http.Server{
						Addr:    fmt.Sprintf(":%d", port),
						Handler: mux,
					}

					lifecycle.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							logger.Info("starting metrics server",
								zap.Int("port", port),
								zap.String("dashboard_url", fmt.Sprintf("http://localhost:%d/dashboard", port)),
								zap.String("metrics_url", fmt.Sprintf("http://localhost:%d/metrics", port)),
								zap.String("api_url", fmt.Sprintf("http://localhost:%d/api/v1/query", port)))
							go func() {
								if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
									logger.Error("server failed", zap.Error(err))
								}
							}()
							return nil
						},
						OnStop: func(ctx context.Context) error {
							logger.Info("stopping server")
							return server.Shutdown(ctx)
						},
					})
				}),
			)

			if err := app.Start(ctx); err != nil {
				return err
			}

			<-stop
			return app.Stop(ctx)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
	cmd.Flags().BoolVarP(&debug, "debug", "v", false, "Enable debug logging")

	return cmd
}

func matchesQuery(metricName, query string) bool {
	// Simple exact match for metric names
	return metricName == query
}
