/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/jonesrussell/goprowl/search/engine"
	"github.com/spf13/cobra"
)

func NewSearchCmd() *cobra.Command {
	var query string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search indexed documents",
		Long:  `Search through crawled and indexed documents using keywords.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(func(searchEngine engine.SearchEngine) error {
				// Create query processor directly from the engine package
				processor := engine.NewQueryProcessor()
				searchQuery, err := processor.ParseQuery(query)
				if err != nil {
					return fmt.Errorf("failed to parse query: %w", err)
				}

				// Set pagination
				searchQuery.SetPagination(1, 10)

				// Perform search
				results, err := searchEngine.Search(searchQuery)
				if err != nil {
					return fmt.Errorf("search failed: %w", err)
				}

				// Display search results

				total := results.Metadata["total"].(int64)
				fmt.Printf("Found %d results:\n\n", total)
				for _, hit := range results.Hits {
					content := hit.Content
					fmt.Printf("Title: %s\n", content["title"])
					fmt.Printf("URL: %s\n", content["url"])
					if snippet, ok := content["snippet"].(string); ok {
						fmt.Printf("Snippet: %s\n", snippet)
					}
					fmt.Printf("Score: %.2f\n", hit.Score)
					fmt.Println("---")
				}

				return nil
			})
		},
	}

	cmd.Flags().StringVarP(&query, "query", "q", "", "Search query (required)")
	cmd.MarkFlagRequired("query")

	return cmd
}
