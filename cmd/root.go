package cmd

import (
	"context"

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
}

func Execute() error {
	app := fx.New(
		app.Module,
		fx.NopLogger,
	)

	if err := app.Start(context.Background()); err != nil {
		return err
	}
	defer app.Stop(context.Background())

	return rootCmd.Execute()
}

var Module = fx.Module("root",
	app.Module,
)

func Run(fn interface{}) error {
	app := fx.New(
		Module,
		fx.Invoke(fn),
		fx.NopLogger,
	)

	if err := app.Start(context.Background()); err != nil {
		return err
	}
	defer app.Stop(context.Background())

	return nil
}
