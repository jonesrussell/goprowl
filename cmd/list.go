/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all indexed documents",
	Long:  `Display all documents currently stored in the index.`,
	Run: func(cmd *cobra.Command, args []string) {
		app := fx.New(
			StorageModule,
			EngineModule,
			fx.Provide(NewApplication),
			fx.Invoke(func(app *Application) {
				if err := app.ListDocuments(); err != nil {
					log.Printf("Error listing documents: %v", err)
				}
			}),
		)
		app.Run()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
