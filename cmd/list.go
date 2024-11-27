/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/jonesrussell/goprowl/internal/app"
	"github.com/jonesrussell/goprowl/search/engine"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all indexed documents",
		Long:  `Display a list of all documents that have been crawled and indexed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fx.New(
				app.Module,
				fx.Invoke(func(engine engine.SearchEngine) error {
					docs, err := engine.List()
					if err != nil {
						return fmt.Errorf("failed to list documents: %w", err)
					}

					displayDocuments(docs)
					return nil
				}),
				fx.NopLogger,
			).Start(cmd.Context())
		},
	}

	return cmd
}

func displayDocuments(docs []engine.Document) {
	fmt.Printf("Found %d documents:\n\n", len(docs))
	for _, doc := range docs {
		content := doc.Content()
		fmt.Printf("Title: %s\n", content["title"])
		fmt.Printf("URL: %s\n", content["url"])
		fmt.Printf("Type: %s\n", doc.Type())
		if createdAt, ok := doc.Metadata()["created_at"]; ok {
			fmt.Printf("Created: %s\n", createdAt)
		}
		fmt.Println("---")
	}
}
