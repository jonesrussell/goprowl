/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jonesrussell/goprowl/internal/app"
	"github.com/jonesrussell/goprowl/metrics"
	"github.com/jonesrussell/goprowl/search/crawlers"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
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
	cmd.MarkFlagRequired("url")

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
	return fx.New(
		app.Module,
		fx.Supply(&crawlers.ConfigOptions{
			URL:      opts.url,
			MaxDepth: opts.depth,
			Debug:    opts.debug,
		}),
		fx.Supply(metrics.Config{
			PushgatewayURL: "http://localhost:9091",
		}),
		metrics.Module,
		crawlers.Module,
		fx.Invoke(func(lc fx.Lifecycle, shutdowner fx.Shutdowner, crawler crawlers.Crawler) error {
			ctx, cancel := context.WithCancel(context.Background())

			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					go func() {
						defer cancel()
						if err := crawler.Crawl(ctx, opts.url, opts.depth); err != nil {
							_ = shutdowner.Shutdown()
							return
						}
						_ = shutdowner.Shutdown()
					}()
					return nil
				},
				OnStop: func(context.Context) error {
					cancel()
					return nil
				},
			})
			return nil
		}),
	)
}
