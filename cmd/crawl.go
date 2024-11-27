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

func NewCrawlCmd() *cobra.Command {
	var url string
	var depth int
	var debug bool

	cmd := &cobra.Command{
		Use:   "crawl",
		Short: "Crawl a website",
		RunE: func(cmd *cobra.Command, args []string) error {
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
				fx.Invoke(func(crawler crawlers.Crawler) error {
					fmt.Printf("Starting crawl of %s with depth %d\n", url, depth)
					return crawler.Crawl(cmd.Context(), url, depth)
				}),
			)

			startCtx := cmd.Context()
			if err := app.Start(startCtx); err != nil {
				return err
			}

			// Ensure we stop the application when we're done
			defer func() {
				stopCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				if err := app.Stop(stopCtx); err != nil {
					fmt.Printf("Error stopping application: %v\n", err)
				}
			}()

			<-app.Done()
			return nil
		},
	}

	cmd.Flags().StringVarP(&url, "url", "u", "", "Starting URL for crawling (required)")
	cmd.Flags().IntVarP(&depth, "depth", "d", 1, "Maximum crawl depth")
	cmd.Flags().BoolVarP(&debug, "debug", "v", false, "Enable debug logging")
	cmd.MarkFlagRequired("url")

	return cmd
}
