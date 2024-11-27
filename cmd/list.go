/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/jonesrussell/goprowl/internal/app"
	"github.com/jonesrussell/goprowl/metrics"
	"github.com/jonesrussell/goprowl/search/engine"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

type ListOptions struct {
	format string
	debug  bool
}

func NewListCmd() *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all indexed documents",
		Long: `Display a list of all documents that have been crawled and indexed.
Supports multiple output formats and filtering options.

Examples:
  goprowl list                  # List all documents in table format
  goprowl list --format json    # Output in JSON format
  goprowl list --format simple  # Simple text output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context(), opts)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&opts.format, "format", "f", "table", "Output format (table, json, simple)")
	cmd.Flags().BoolVarP(&opts.debug, "debug", "v", false, "Enable debug output")

	return cmd
}

func runList(ctx context.Context, opts *ListOptions) error {
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
		fx.Invoke(func(engine engine.SearchEngine, logger *zap.Logger, metrics *metrics.ComponentMetrics) error {
			startTime := time.Now()
			defer func() {
				metrics.ObserveHistogram(
					"list_documents_duration_seconds",
					time.Since(startTime).Seconds(),
				)
			}()

			logger.Info("listing documents")
			docs, err := engine.List()
			if err != nil {
				metrics.IncCounter("list_documents_errors_total", 1)
				return fmt.Errorf("failed to list documents: %w", err)
			}

			metrics.SetGaugeValue("indexed_documents_total", float64(len(docs)))
			return displayDocuments(docs, opts.format)
		}),
	}

	if !opts.debug {
		options = append(options, fx.NopLogger)
	}

	return fx.New(options...).Start(ctx)
}

func displayDocuments(docs []engine.Document, format string) error {
	switch format {
	case "json":
		return displayJSON(docs)
	case "table":
		return displayTable(docs)
	case "simple":
		return displaySimple(docs)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func displayJSON(docs []engine.Document) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(docs)
}

func displayTable(docs []engine.Document) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "URL\tTitle\tType\tCreated\n")
	fmt.Fprintf(w, "---\t-----\t----\t-------\n")

	for _, doc := range docs {
		content := doc.Content()
		metadata := doc.Metadata()
		fmt.Fprintf(w, "%s\t%s\t%s\t%v\n",
			content["url"],
			content["title"],
			doc.Type(),
			metadata["created_at"],
		)
	}
	return w.Flush()
}

func displaySimple(docs []engine.Document) error {
	for _, doc := range docs {
		content := doc.Content()
		fmt.Printf("%s - %s\n", content["title"], content["url"])
	}
	return nil
}
