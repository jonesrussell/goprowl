package crawlers

import (
	"go.uber.org/fx"
)

// Module provides crawler dependencies
var Module = fx.Module("crawlers",
	fx.Provide(
		fx.Annotate(
			NewCollyCrawler,
			fx.As(new(Crawler)),
		),
	),
)
