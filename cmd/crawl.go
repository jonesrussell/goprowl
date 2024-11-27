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

func NewCrawlCmd() *cobra.Command {
	var url string
	var depth int

	cmd := &cobra.Command{
		Use:   "crawl",
		Short: "Crawl a website",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fx.New(
				fx.Supply(&crawlers.ConfigOptions{
					URL:      url,
					MaxDepth: depth,
				}),
				app.Module,
				fx.Invoke(func(crawler crawlers.Crawler) error {
					fmt.Printf("Starting crawl of %s with depth %d\n", url, depth)
					return crawler.Crawl(cmd.Context(), url, depth)
				}),
			).Start(cmd.Context())
		},
	}

	cmd.Flags().StringVarP(&url, "url", "u", "", "Starting URL for crawling (required)")
	cmd.Flags().IntVarP(&depth, "depth", "d", 1, "Maximum crawl depth")
	cmd.MarkFlagRequired("url")

	return cmd
}
