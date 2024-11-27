package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

			app := fx.New(
				NewLoggerModule(),
				app.Module,
				metrics.Module,
				fx.Invoke(func(lifecycle fx.Lifecycle, logger *zap.Logger, registry *prometheus.Registry) {
					mux := http.NewServeMux()

					// Register all handlers
					metrics.RegisterDashboard(mux)
					mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
					mux.HandleFunc("/api/v1/query", func(w http.ResponseWriter, r *http.Request) {
						logger.Info("Received query request", zap.String("url", r.URL.String()))

						query := r.URL.Query().Get("query")
						if query == "" {
							http.Error(w, "query parameter is required", http.StatusBadRequest)
							return
						}

						_, err := registry.Gather()
						if err != nil {
							logger.Error("Failed to gather metrics", zap.Error(err))
							http.Error(w, "internal server error", http.StatusInternalServerError)
							return
						}

						// Process and return metrics
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
								Result: make([]struct {
									Metric map[string]string `json:"metric"`
									Value  []interface{}     `json:"value"`
								}, 0),
							},
						}

						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(response)
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

			if err := app.Start(cmd.Context()); err != nil {
				return err
			}

			<-stop
			return app.Stop(cmd.Context())
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
	cmd.Flags().BoolVarP(&debug, "debug", "v", false, "Enable debug logging")

	return cmd
}
