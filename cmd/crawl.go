/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	startURL string
	maxDepth int
)

var crawlCmd = &cobra.Command{
	Use:   "crawl",
	Short: "Start crawling from a given URL",
	Long: `Crawl starts from the specified URL and indexes all pages 
found within the maximum depth limit.`,
	Run: func(cmd *cobra.Command, args []string) {
		app := fx.New(
			StorageModule,
			EngineModule,
			CrawlerModule,
			fx.Provide(
				func() *Config {
					return &Config{
						StartURL: startURL,
						MaxDepth: maxDepth,
					}
				},
				NewApplication,
			),
			fx.Invoke(runCrawl),
		)
		app.Run()
	},
}

func init() {
	rootCmd.AddCommand(crawlCmd)
	crawlCmd.Flags().StringVarP(&startURL, "url", "u", "https://go.dev", "The URL to start crawling from")
	crawlCmd.Flags().IntVarP(&maxDepth, "depth", "d", 3, "Maximum crawl depth")
}
