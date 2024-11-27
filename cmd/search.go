/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/jonesrussell/goprowl/internal/app"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var query string

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search indexed documents",
	Long: `Search through indexed documents using various query types:
- Phrase matching: "exact phrase"
- Boolean operators: term1 AND term2, NOT term
- Fuzzy matching: word~2
- Field search: title:word`,
	Run: func(cmd *cobra.Command, args []string) {
		fxApp := fx.New(
			app.StorageModule,
			app.EngineModule,
			app.CrawlerModule,
			fx.Provide(
				func() *app.Config {
					return &app.Config{}
				},
				app.NewApplication,
			),
			fx.Invoke(func(application *app.Application) {
				if err := application.Search(query); err != nil {
					log.Printf("Error searching documents: %v", err)
				}
			}),
		)

		fxApp.Run()
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringVarP(&query, "query", "q", "", "Search query")
	searchCmd.MarkFlagRequired("query")
}
