package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonesrussell/goprowl/internal/logger"
	"github.com/jonesrussell/goprowl/metrics"
	"github.com/jonesrussell/goprowl/search/crawlers"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

var (
	rootCmd = &cobra.Command{
		Use:   "goprowl",
		Short: "GoProwl is a web crawler and search engine",
		Long: `A flexible web crawler and search engine built with Go 
that supports full-text search, concurrent crawling, and 
configurable storage backends.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	globalLogger logger.Logger
)

// GetRootCmd returns the root command instance
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// NewLoggerModule provides the application-wide logger
func NewLoggerModule() fx.Option {
	return fx.Module("logger",
		fx.Provide(
			func() (logger.Logger, error) {
				// Check if debug flag is set via cobra command
				debug := false
				if cmd := GetRootCmd(); cmd != nil {
					debug, _ = cmd.Flags().GetBool("debug")
				}

				var zapLogger *zap.Logger
				var err error
				if debug {
					zapLogger, err = zap.NewDevelopment()
				} else {
					zapLogger, err = zap.NewProduction()
				}
				if err != nil {
					return nil, fmt.Errorf("failed to create logger: %w", err)
				}

				return logger.NewZapLogger(zapLogger), nil
			},
		),
	)
}

func Execute() error {
	app := fx.New(
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{
				Logger: zap.L(),
			}
		}),

		NewLoggerModule(),
		metrics.Module,
		crawlers.Module,

		fx.Invoke(func(l logger.Logger) {
			globalLogger = l
		}),
	)

	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	signalCtx, signalCancel := context.WithCancel(context.Background())
	defer signalCancel()

	go func() {
		select {
		case sig := <-sigChan:
			globalLogger.Info(context.Background(), "received signal", logger.Field{Key: "signal", Value: sig.String()})
			cancel()

			select {
			case sig := <-sigChan:
				globalLogger.Fatal(context.Background(), "received signal", logger.Field{Key: "signal", Value: sig.String()})
			case <-time.After(10 * time.Second):
				globalLogger.Fatal(context.Background(), "graceful shutdown timed out, force quitting")
			}
		case <-signalCtx.Done():
			return
		}
	}()

	rootCmd.AddCommand(
		NewCrawlCmd(),
		NewSearchCmd(),
		NewListCmd(),
	)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		signalCancel()
		globalLogger.Error(context.Background(), "error occurred", logger.Field{Key: "error", Value: err.Error()})
		return fmt.Errorf("execution error: %w", err)
	}

	signalCancel()
	globalLogger.Info(context.Background(), "application completed successfully")
	return nil
}

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
