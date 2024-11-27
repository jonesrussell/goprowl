/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/jonesrussell/goprowl/search/engine"
	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all indexed documents",
		Long:  `Display a list of all documents that have been crawled and indexed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(func(engine engine.SearchEngine) error {
				docs, err := engine.List()
				if err != nil {
					return fmt.Errorf("failed to list documents: %w", err)
				}

				// Format and display documents
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
				return nil
			})
		},
	}

	return cmd
}
