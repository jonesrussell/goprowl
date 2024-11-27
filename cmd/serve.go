package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonesrussell/goprowl/internal/app"
	"github.com/jonesrussell/goprowl/metrics"
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
			// Create a channel to wait for interrupt signal
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

			app := fx.New(
				NewLoggerModule(),
				app.Module,
				metrics.Module,
				fx.Invoke(func(lifecycle fx.Lifecycle, logger *zap.Logger) {
					mux := http.NewServeMux()
					metrics.RegisterDashboard(mux)

					server := &http.Server{
						Addr:    fmt.Sprintf(":%d", port),
						Handler: mux,
					}

					lifecycle.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							logger.Info("starting dashboard server",
								zap.Int("port", port),
								zap.String("url", fmt.Sprintf("http://localhost:%d/dashboard", port)))
							go func() {
								if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
									logger.Error("dashboard server failed", zap.Error(err))
								}
							}()
							return nil
						},
						OnStop: func(ctx context.Context) error {
							logger.Info("stopping dashboard server")
							return server.Shutdown(ctx)
						},
					})
				}),
			)

			if err := app.Start(cmd.Context()); err != nil {
				return err
			}

			// Wait for interrupt signal
			<-stop

			// Gracefully shutdown
			return app.Stop(cmd.Context())
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the dashboard server on")
	cmd.Flags().BoolVarP(&debug, "debug", "v", false, "Enable debug logging")

	return cmd
}
