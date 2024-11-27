/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/jonesrussell/goprowl/internal/app"
	"github.com/jonesrussell/goprowl/search/crawlers"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type crawlOptions struct {
	url   string
	depth int
}

func NewCrawlCmd() *cobra.Command {
	opts := &crawlOptions{}

	cmd := &cobra.Command{
		Use:   "crawl",
		Short: "Start crawling from a given URL",
		Long:  `Crawl a website starting from the specified URL up to the given depth.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fx.New(
				app.Module,
				fx.Invoke(func(crawler crawlers.Crawler) error {
					fmt.Printf("Starting crawl of %s with depth %d\n", opts.url, opts.depth)
					return crawler.Crawl(cmd.Context(), opts.url, opts.depth)
				}),
				fx.NopLogger,
			).Start(cmd.Context())
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&opts.url, "url", "u", "", "Starting URL for crawling (required)")
	cmd.Flags().IntVarP(&opts.depth, "depth", "d", 1, "Maximum crawl depth")
	cmd.MarkFlagRequired("url")

	return cmd
}
