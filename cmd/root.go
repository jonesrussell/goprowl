package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   "goprowl",
	Short: "GoProwl is a web crawler and search engine",
	Long: `A flexible web crawler and search engine built with Go 
that supports full-text search, concurrent crawling, and 
configurable storage backends.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var logger *zap.Logger

func NewLoggerModule() fx.Option {
	return fx.Module("logger",
		fx.Provide(
			func() (*zap.Logger, error) {
				// Configure production logger
				logger, err := zap.NewProduction()
				if err != nil {
					return nil, fmt.Errorf("failed to create logger: %w", err)
				}
				return logger, nil
			},
		),
	)
}

func Execute() error {
	// Create fx application
	app := fx.New(
		NewLoggerModule(),
		fx.Invoke(func(log *zap.Logger) {
			logger = log // Store logger in package variable
		}),
	)

	// Start fx application
	if err := app.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}
	defer app.Stop(context.Background())

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
			logger.Info("received signal, initiating graceful shutdown",
				zap.String("signal", sig.String()))
			cancel()

			select {
			case sig := <-sigChan:
				logger.Fatal("received second signal, force quitting",
					zap.String("signal", sig.String()))
			case <-time.After(10 * time.Second):
				logger.Fatal("graceful shutdown timed out, force quitting")
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
		logger.Error("execution error", zap.Error(err))
		return fmt.Errorf("execution error: %w", err)
	}

	signalCancel() // Clean up signal handler
	logger.Info("application completed successfully")
	return nil
}
