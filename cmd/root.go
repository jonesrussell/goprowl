package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
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
	// Create a cancellable context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create buffered channel for signals
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Handle interrupt signals in a separate context
	signalCtx, signalCancel := context.WithCancel(context.Background())
	defer signalCancel()

	go func() {
		select {
		case sig := <-sigChan:
			fmt.Printf("\nReceived signal %v. Initiating graceful shutdown...\n", sig)
			// Cancel the main context
			cancel()

			// Wait for second signal for force quit
			select {
			case sig := <-sigChan:
				fmt.Printf("\nReceived second signal %v. Force quitting...\n", sig)
				os.Exit(1)
			case <-time.After(10 * time.Second):
				fmt.Println("\nGraceful shutdown timed out. Force quitting...")
				os.Exit(1)
			}
		case <-signalCtx.Done():
			return
		}
	}()

	// Add commands
	rootCmd.AddCommand(
		NewCrawlCmd(),
		NewSearchCmd(),
		NewListCmd(),
	)

	// Execute with context and handle any errors
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		signalCancel() // Clean up signal handler
		return fmt.Errorf("execution error: %w", err)
	}

	signalCancel() // Clean up signal handler
	return nil
}
