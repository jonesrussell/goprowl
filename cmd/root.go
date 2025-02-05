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
)

// GetRootCmd returns the root command instance
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// LoggerModule provides the application-wide logger
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

				return logger.NewZapLogger(zapLogger), nil // Return the custom logger
			},
		),
	)
}

func Execute() error {
	// Create fx application with all required modules
	app := fx.New(
		// Configure logging - reduce fx verbosity
		fx.WithLogger(func(log logger.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{
				Logger: log, // Use the custom logger
			}
		}),

		// Add the logger module
		NewLoggerModule(),

		// Add other modules that depend on the logger
		metrics.Module,
		crawlers.Module,

		// Configure error handling
		fx.NopLogger,
	)

	// Start the application
	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var globalLogger logger.Logger
	if err := app.Start(startCtx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	// Retrieve the logger from the app
	if err := app.Invoke(func(l logger.Logger) {
		globalLogger = l
	}); err != nil {
		return fmt.Errorf("failed to retrieve logger: %w", err)
	}

	// Create a cancellable context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create buffered channel for signals
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Handle interrupt signals in a separate context
	signalCtx, signalCancel := context.WithCancel(context.Background())
	defer signalCancel()

	// Update signal handling to use logger
	go func() {
		select {
		case sig := <-sigChan:
			globalLogger.Info("received signal, initiating graceful shutdown", zap.String("signal", sig.String()))
			cancel()

			select {
			case sig := <-sigChan:
				globalLogger.Fatal("received second signal, force quitting", zap.String("signal", sig.String()))
			case <-time.After(10 * time.Second):
				globalLogger.Fatal("graceful shutdown timed out, force quitting")
			}
		case <-signalCtx.Done():
			return
		}
	}()

	// Add commands
	rootCmd.AddCommand(
		NewCrawlCmd(),
		NewSearchCmd(),
		NewListCmd(),
	)

	// Execute with context and handle any errors
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		signalCancel() // Clean up signal handler
		globalLogger.Error("execution error", zap.Error(err))
		return fmt.Errorf("execution error: %w", err)
	}

	signalCancel() // Clean up signal handler
	globalLogger.Info("application completed successfully")
	return nil
}
