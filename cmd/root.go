package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonesrussell/goprowl/internal/app"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var rootCmd = &cobra.Command{
	Use:   "goprowl",
	Short: "GoProwl is a web crawler and search engine",
	Long: `A flexible web crawler and search engine built with Go 
that supports full-text search, concurrent crawling, and 
configurable storage backends.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() error {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal. Shutting down...")
		cancel()
	}()

	// Create the application
	application := fx.New(
		app.Module,
		fx.NopLogger,
	)

	// Start the application
	if err := application.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}
	defer func() {
		if err := application.Stop(ctx); err != nil {
			fmt.Printf("Error stopping application: %v\n", err)
		}
	}()

	// Add commands
	rootCmd.AddCommand(
		NewCrawlCmd(),
		NewSearchCmd(),
		NewListCmd(),
	)

	// Execute the root command
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return err
	}

	return nil
}

// Run executes a function with dependency injection
func Run(fn interface{}) error {
	application := fx.New(
		app.Module,
		fx.Invoke(fn),
		fx.NopLogger,
	)

	// Start the application with context
	ctx := context.Background()
	if err := application.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	// Ensure application is stopped
	defer func() {
		if err := application.Stop(ctx); err != nil {
			fmt.Printf("Error stopping application: %v\n", err)
		}
	}()

	return nil
}

var Module = fx.Module("root",
	app.Module,
)
