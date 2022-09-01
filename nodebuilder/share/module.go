package share

import (
	"context"

	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/service/share"
)

func Module(tp node.Type, cfg *Config, options ...fx.Option) fx.Option {
	baseComponents := fx.Options(
		fx.Supply(cfg),
		fx.Options(options...),
		fx.Invoke(cfg.ValidateBasic),
		fx.Invoke(share.EnsureEmptySquareExists),
		fx.Provide(Discovery(*cfg)),
		fx.Provide(fx.Annotate(
			share.NewService,
			fx.OnStart(func(ctx context.Context, service *share.Service) error {
				return service.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, service *share.Service) error {
				return service.Stop(ctx)
			}),
		)),
	)

	switch tp {
	case node.Light:
		return fx.Module(
			"share",
			baseComponents,
			fx.Provide(fx.Annotate(
				share.NewLightAvailability,
				fx.OnStart(func(ctx context.Context, avail *share.LightAvailability) error {
					return avail.Start(ctx)
				}),
				fx.OnStop(func(ctx context.Context, avail *share.LightAvailability) error {
					return avail.Stop(ctx)
				}),
			)),
			// CacheAvailability's lifecycle continues to use a fx hook,
			// since the LC requires a CacheAvailability but the constructor returns a share.Availability
			fx.Provide(CacheAvailability[*share.LightAvailability]),
		)
	case node.Bridge, node.Full:
		return fx.Module(
			"share",
			baseComponents,
			fx.Provide(fx.Annotate(
				share.NewFullAvailability,
				fx.OnStart(func(ctx context.Context, avail *share.FullAvailability) error {
					return avail.Start(ctx)
				}),
				fx.OnStop(func(ctx context.Context, avail *share.FullAvailability) error {
					return avail.Stop(ctx)
				}),
			)),
			// CacheAvailability's lifecycle continues to use a fx hook,
			// since the LC requires a CacheAvailability but the constructor returns a share.Availability
			fx.Provide(CacheAvailability[*share.FullAvailability]),
		)
	default:
		panic("invalid node type")
	}
}
