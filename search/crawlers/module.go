package crawlers

import (
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"go.uber.org/fx"
)

var Module = fx.Module("crawlers",
	fx.Provide(
		func() *colly.Collector {
			c := colly.NewCollector(
				colly.Debugger(&debug.LogDebugger{}),
			)
			return c
		},
		NewConfig,
		fx.Annotate(
			NewCollyCrawler,
			fx.As(new(Crawler)),
		),
	),
)
