/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jonesrussell/goprowl/internal/app"
	"github.com/jonesrussell/goprowl/internal/logger"
	"github.com/jonesrussell/goprowl/metrics"
	"github.com/jonesrussell/goprowl/search/adapters/storage"
	"github.com/jonesrussell/goprowl/search/crawlers"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// CrawlOptions holds the command-line options for the crawl command
type CrawlOptions struct {
	url   string
	depth int
	debug bool
}

// NewCrawlCmd creates the 'crawl' command.
func NewCrawlCmd() *cobra.Command {
	opts := &CrawlOptions{}

	cmd := &cobra.Command{
		Use:   "crawl",
		Short: "Crawl a website",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCrawl(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.url, "url", "u", "", "Starting URL for crawling (required)")
	cmd.Flags().IntVarP(&opts.depth, "depth", "d", 1, "Maximum crawl depth")
	cmd.Flags().BoolVarP(&opts.debug, "debug", "v", false, "Enable debug logging")
	if err := cmd.MarkFlagRequired("url"); err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("failed to mark flag as required: %w", err)
		}
	}

	return cmd
}

// runCrawl handles the main crawl command execution
func runCrawl(ctx context.Context, opts *CrawlOptions) error {
	app := createApp(opts)

	startCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	<-app.Done()

	if err := app.Err(); err != nil {
		return fmt.Errorf("application error: %w", err)
	}

	return nil
}

// createApp initializes the fx application with the necessary modules and config.
func createApp(opts *CrawlOptions) *fx.App {
	// Set logging level based on debug flag
	logLevel := logger.InfoLevel
	if opts.debug {
		logLevel = logger.DebugLevel
	}

	options := []fx.Option{
		fx.WithLogger(func(log *logger.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{
				Logger: log.WithOptions(logger.IncreaseLevel(logLevel)),
			}
		}),
		NewLoggerModule(),
		fx.Provide(
			func() *crawlers.ConfigOptions {
				return &crawlers.ConfigOptions{
					URL:      opts.url,
					MaxDepth: opts.depth,
					Debug:    opts.debug,
				}
			},
		),
		metrics.Module,
		app.Module,
		crawlers.Module,
		storage.Module,
		// Add lifecycle hook to handle crawler completion
		fx.Invoke(func(
			lifecycle fx.Lifecycle,
			shutdowner fx.Shutdowner,
			crawler crawlers.Crawler,
			storageAdapter *storage.StorageAdapter,
			logger *logger.Logger,
		) error {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("starting crawler", logger.NewField("url", opts.url), logger.NewField("depth", opts.depth))

					go func() {
						if err := crawler.CrawlWithHandler(ctx, opts.url, opts.depth, storageAdapter.HandleCrawledPage); err != nil {
							logger.Error("crawler failed", logger.NewField("error", err))
						}

						if err := shutdowner.Shutdown(); err != nil {
							logger.Error("shutdown failed", logger.NewField("error", err))
						}
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("stopping crawler")
					return nil
				},
			})
			return nil
		}),
	}

	// Only add NopLogger in non-debug mode
	if !opts.debug {
		options = append(options, fx.NopLogger)
	}

	return fx.New(options...)
}
