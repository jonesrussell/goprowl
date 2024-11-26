/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

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
		app := fx.New(
			StorageModule,
			EngineModule,
			fx.Provide(NewApplication),
			fx.Invoke(func(app *Application) {
				if err := app.Search(query); err != nil {
					log.Printf("Error searching documents: %v", err)
				}
			}),
		)
		app.Run()
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringVarP(&query, "query", "q", "", "Search query")
	searchCmd.MarkFlagRequired("query")
}
