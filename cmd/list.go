/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/jonesrussell/goprowl/internal/app"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all indexed documents",
	Long:  `Display all documents currently stored in the index.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return Run(func(application *app.Application) error {
			return application.ListDocuments()
		})
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
