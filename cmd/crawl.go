/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"

	"github.com/jonesrussell/goprowl/internal/app"
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
		fxApp := fx.New(
			app.StorageModule,
			app.EngineModule,
			app.CrawlerModule,
			fx.Provide(
				func() *app.Config {
					return &app.Config{
						StartURL: startURL,
						MaxDepth: maxDepth,
					}
				},
				app.NewApplication,
			),
			fx.Invoke(func(application *app.Application) {
				if err := application.Run(context.Background()); err != nil {
					log.Printf("Error running application: %v", err)
				}
			}),
		)
		fxApp.Run()
	},
}

func init() {
	rootCmd.AddCommand(crawlCmd)
	crawlCmd.Flags().StringVarP(&startURL, "url", "u", "https://go.dev", "The URL to start crawling from")
	crawlCmd.Flags().IntVarP(&maxDepth, "depth", "d", 3, "Maximum crawl depth")
}
