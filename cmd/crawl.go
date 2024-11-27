/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"

	"github.com/jonesrussell/goprowl/search/crawlers"
	"github.com/spf13/cobra"
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
			return Run(func(crawler crawlers.Crawler) error {
				return crawler.Crawl(context.Background(), opts.url, opts.depth)
			})
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&opts.url, "url", "u", "", "Starting URL for crawling (required)")
	cmd.Flags().IntVarP(&opts.depth, "depth", "d", 1, "Maximum crawl depth")
	cmd.MarkFlagRequired("url")

	return cmd
}
