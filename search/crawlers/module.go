package crawlers

import (
	"go.uber.org/fx"
)

var Module = fx.Module("crawlers",
	fx.Provide(
		NewConfig,
		fx.Annotate(
			NewCrawlerFromConfig,
			fx.As(new(Crawler)),
		),
	),
)
