package crawlers

import (
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"go.uber.org/fx"
)

var Module = fx.Module("crawlers",
	fx.Provide(
		func(cfg *Config) *colly.Collector {
			opts := []colly.CollectorOption{}

			// Only add debug logger if debug mode is enabled
			if cfg.Debug {
				opts = append(opts, colly.Debugger(&debug.LogDebugger{}))
			}

			return colly.NewCollector(opts...)
		},
		NewConfig,
		fx.Annotate(
			NewCollyCrawler,
			fx.As(new(Crawler)),
		),
	),
)
