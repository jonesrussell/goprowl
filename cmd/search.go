/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jonesrussell/goprowl/internal/app"
	"github.com/jonesrussell/goprowl/metrics"
	"github.com/jonesrussell/goprowl/search/engine"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

type SearchOptions struct {
	query     string
	page      int
	limit     int
	debug     bool
	format    string
	sortBy    string
	sortOrder string
}

func NewSearchCmd() *cobra.Command {
	opts := &SearchOptions{}

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search indexed documents",
		Long: `Search through crawled and indexed documents using keywords.
Supports pagination, sorting, and multiple output formats.

Examples:
  goprowl search -q "golang programming"              # Basic search
  goprowl search -q "web crawler" --page 2            # Paginated search
  goprowl search -q "api" --limit 20                  # Custom result limit
  goprowl search -q "test" --format json              # JSON output
  goprowl search -q "docs" --sort-by score --desc     # Sort by relevance`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(cmd.Context(), opts)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&opts.query, "query", "q", "", "Search query (required)")
	cmd.Flags().IntVarP(&opts.page, "page", "p", 1, "Page number")
	cmd.Flags().IntVarP(&opts.limit, "limit", "l", 10, "Results per page")
	cmd.Flags().BoolVarP(&opts.debug, "debug", "v", false, "Enable debug output")
	cmd.Flags().StringVarP(&opts.format, "format", "f", "text", "Output format (text, json, table)")
	cmd.Flags().StringVar(&opts.sortBy, "sort-by", "score", "Sort results by (score, date, title)")
	cmd.Flags().StringVar(&opts.sortOrder, "sort-order", "desc", "Sort order (asc, desc)")

	cmd.MarkFlagRequired("query")

	return cmd
}

func runSearch(ctx context.Context, opts *SearchOptions) error {
	logLevel := zap.InfoLevel
	if opts.debug {
		logLevel = zap.DebugLevel
	}

	options := []fx.Option{
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Provide(func() (*zap.Logger, error) {
			config := zap.NewProductionConfig()
			config.Level = zap.NewAtomicLevelAt(logLevel)
			return config.Build()
		}),
		app.Module,
		metrics.Module,
		fx.Invoke(func(searchEngine engine.SearchEngine, logger *zap.Logger, metrics *metrics.ComponentMetrics) error {
			startTime := time.Now()
			defer func() {
				metrics.ObserveHistogram(
					"search_duration_seconds",
					time.Since(startTime).Seconds(),
				)
			}()

			logger.Info("executing search",
				zap.String("query", opts.query),
				zap.Int("page", opts.page),
				zap.Int("limit", opts.limit),
				zap.String("sort_by", opts.sortBy),
				zap.String("sort_order", opts.sortOrder))

			results, err := executeSearch(searchEngine, opts)
			if err != nil {
				metrics.IncCounter("search_errors_total", 1)
				logger.Error("search failed", zap.Error(err))
				return fmt.Errorf("search failed: %w", err)
			}

			metrics.SetGaugeValue("search_results_total", float64(results.Metadata["total"].(int64)))
			return displayResults(results, opts.format)
		}),
	}

	if !opts.debug {
		options = append(options, fx.NopLogger)
	}

	return fx.New(options...).Start(ctx)
}

func executeSearch(searchEngine engine.SearchEngine, opts *SearchOptions) (*engine.SearchResults, error) {
	processor := engine.NewQueryProcessor()
	searchQuery, err := processor.ParseQuery(opts.query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	searchQuery.SetPagination(opts.page, opts.limit)
	searchQuery.SortField = opts.sortBy
	searchQuery.SortDescending = strings.EqualFold(opts.sortOrder, "desc")

	return searchEngine.Search(searchQuery)
}

func displayResults(results *engine.SearchResults, format string) error {
	switch format {
	case "json":
		return displayJSONResults(results)
	case "table":
		return displayTableResults(results)
	default:
		return displayTextResults(results)
	}
}

func displayTextResults(results *engine.SearchResults) error {
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
		if date, ok := content["date"].(time.Time); ok {
			fmt.Printf("Date: %s\n", date.Format(time.RFC3339))
		}
		fmt.Println("---")
	}
	return nil
}

func displayJSONResults(results *engine.SearchResults) error {
	// Implementation similar to list command's JSON display
	return nil // TODO: Implement
}

func displayTableResults(results *engine.SearchResults) error {
	// Implementation similar to list command's table display
	return nil // TODO: Implement
}
