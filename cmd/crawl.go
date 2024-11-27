/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"os"
	"time"

	"github.com/jonesrussell/goprowl/internal/app"
	"github.com/jonesrussell/goprowl/metrics"
	"github.com/jonesrussell/goprowl/search/crawlers"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func NewCrawlCmd() *cobra.Command {
	var url string
	var depth int
	var debug bool

	cmd := &cobra.Command{
		Use:   "crawl",
		Short: "Crawl a website",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := &fxevent.ConsoleLogger{W: os.Stdout}
			app := fx.New(
				fx.Supply(&crawlers.ConfigOptions{
					URL:      url,
					MaxDepth: depth,
					Debug:    debug,
				}),
				fx.Supply(metrics.Config{
					PushgatewayURL: "http://localhost:9091",
				}),
				metrics.Module,
				app.Module,
				crawlers.Module,
				fx.Invoke(func(lc fx.Lifecycle, shutdowner fx.Shutdowner, crawler crawlers.Crawler) error {
					// Create a cancellable context for the crawler
					ctx, cancel := context.WithCancel(context.Background())

					lc.Append(fx.Hook{
						OnStart: func(context.Context) error {
							logger.LogEvent(&fxevent.OnStartExecuting{
								FunctionName: "Crawler.Start",
								CallerName:   "CrawlCommand",
							})

							// Create a channel to signal crawl completion
							done := make(chan struct{})

							// Run crawl in a goroutine so we can handle shutdown properly
							go func() {
								defer close(done)
								if err := crawler.Crawl(ctx, url, depth); err != nil {
									logger.LogEvent(&fxevent.OnStopExecuted{
										FunctionName: "Crawler.Crawl",
										CallerName:   "CrawlCommand",
										Err:          err,
									})
									return
								}
								logger.LogEvent(&fxevent.OnStopExecuted{
									FunctionName: "Crawler.Crawl",
									CallerName:   "CrawlCommand",
									Runtime:      time.Duration(0),
								})
							}()

							// Start shutdown sequence after crawl completes
							go func() {
								<-done
								logger.LogEvent(&fxevent.OnStopExecuting{
									FunctionName: "Crawler.Shutdown",
									CallerName:   "CrawlCommand",
								})

								// Give a small delay to ensure logs are written
								time.Sleep(100 * time.Millisecond)

								if err := shutdowner.Shutdown(); err != nil {
									logger.LogEvent(&fxevent.OnStopExecuted{
										FunctionName: "Crawler.Shutdown",
										CallerName:   "CrawlCommand",
										Err:          err,
									})
									cancel() // Cancel the context to ensure complete shutdown
									return
								}

								logger.LogEvent(&fxevent.OnStopExecuted{
									FunctionName: "Crawler.Shutdown",
									CallerName:   "CrawlCommand",
									Runtime:      time.Duration(0),
								})
								cancel() // Cancel the context to ensure complete shutdown
							}()

							logger.LogEvent(&fxevent.OnStartExecuted{
								FunctionName: "Crawler.Start",
								CallerName:   "CrawlCommand",
								Runtime:      time.Duration(0),
							})
							return nil
						},
						OnStop: func(context.Context) error {
							logger.LogEvent(&fxevent.OnStopExecuting{
								FunctionName: "Crawler.Stop",
								CallerName:   "CrawlCommand",
							})
							cancel() // Ensure context is cancelled on stop
							return nil
						},
					})
					return nil
				}),
				fx.WithLogger(func() fxevent.Logger {
					return logger
				}),
			)

			startCtx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			if err := app.Start(startCtx); err != nil {
				return err
			}

			// Wait for completion or context cancellation
			select {
			case <-app.Done():
				logger.LogEvent(&fxevent.Stopped{
					Err: app.Err(),
				})
				// Add small delay to ensure final logs are written
				time.Sleep(100 * time.Millisecond)
				return app.Err()
			case <-cmd.Context().Done():
				logger.LogEvent(&fxevent.Stopped{
					Err: cmd.Context().Err(),
				})
				return cmd.Context().Err()
			}
		},
	}

	cmd.Flags().StringVarP(&url, "url", "u", "", "Starting URL for crawling (required)")
	cmd.Flags().IntVarP(&depth, "depth", "d", 1, "Maximum crawl depth")
	cmd.Flags().BoolVarP(&debug, "debug", "v", false, "Enable debug logging")
	cmd.MarkFlagRequired("url")

	return cmd
}
