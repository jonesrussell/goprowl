package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonesrussell/goprowl/internal/app"
	"github.com/jonesrussell/goprowl/metrics"
	"github.com/jonesrussell/goprowl/search/crawlers"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	// Global logger instance
	globalLogger *zap.Logger
)

// GetRootCmd returns the root command instance
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// LoggerModule provides the application-wide logger
func NewLoggerModule() fx.Option {
	return fx.Module("logger",
		fx.Provide(
			func() (*zap.Logger, error) {
				// Check if debug flag is set via cobra command
				debug := false
				if cmd := GetRootCmd(); cmd != nil {
					debug, _ = cmd.Flags().GetBool("debug")
				}

				var config zap.Config
				if debug {
					// Debug configuration with more details
					config = zap.NewDevelopmentConfig()
					config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
					config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
					config.EncoderConfig.TimeKey = "ts"
					config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
				} else {
					// Production configuration
					config = zap.NewProductionConfig()
					config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
				}

				config.OutputPaths = []string{"stdout"}
				config.ErrorOutputPaths = []string{"stderr"}

				logger, err := config.Build()
				if err != nil {
					return nil, fmt.Errorf("failed to create logger: %w", err)
				}
				globalLogger = logger
				zap.ReplaceGlobals(logger)
				return logger, nil
			},
		),
	)
}

func Execute() error {
	// Create base logger for startup
	var err error
	globalLogger, err = zap.NewProduction()
	if err != nil {
		return fmt.Errorf("failed to create startup logger: %w", err)
	}
	defer globalLogger.Sync()

	// Create fx application with all required modules
	app := fx.New(
		// Configure logging - reduce fx verbosity
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			// Create a new logger instead of trying to modify level
			return &fxevent.ZapLogger{
				Logger: log.Named("fx").WithOptions(
					zap.WrapCore(func(core zapcore.Core) zapcore.Core {
						return zapcore.NewCore(
							zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
							zapcore.AddSync(os.Stdout),
							zapcore.WarnLevel,
						)
					}),
				),
			}
		}),

		// Add the logger module
		NewLoggerModule(),

		// Add other modules that depend on the logger
		app.Module,
		metrics.Module,
		crawlers.Module,

		// Configure error handling
		fx.NopLogger,
	)

	// Start the application
	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		globalLogger.Error("failed to start application", zap.Error(err))
		return fmt.Errorf("failed to start application: %w", err)
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
			globalLogger.Info("received signal, initiating graceful shutdown",
				zap.String("signal", sig.String()))
			cancel()

			select {
			case sig := <-sigChan:
				globalLogger.Fatal("received second signal, force quitting",
					zap.String("signal", sig.String()))
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
		NewServeCmd(),
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
