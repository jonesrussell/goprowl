package crawlers

import (
	"go.uber.org/fx"
)

// Module provides crawler dependencies
var Module = fx.Module("crawlers",
	fx.Provide(
		ProvideDefaultConfigOptions,
		NewConfig,
		fx.Annotate(
			NewCrawlerFromConfig,
			fx.As(new(Crawler)),
		),
	),
)
