package pruner

import (
	"context"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/core"
	"github.com/celestiaorg/celestia-node/libs/fxutil"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/pruner"
	"github.com/celestiaorg/celestia-node/pruner/full"
	fullavail "github.com/celestiaorg/celestia-node/share/availability/full"
	"github.com/celestiaorg/celestia-node/share/availability/light"
)

var log = logging.Logger("module/pruner")

func ConstructModule(tp node.Type, cfg *Config) fx.Option {
	baseComponents := fx.Options(
		fx.Supply(cfg),
	)

	prunerService := fx.Options(
		fx.Provide(fx.Annotate(
			newPrunerService,
			fx.OnStart(func(ctx context.Context, p *pruner.Service) error {
				return p.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, p *pruner.Service) error {
				return p.Stop(ctx)
			}),
		)),
		// This is necessary to invoke the pruner service as independent thanks to a
		// quirk in FX.
		fx.Invoke(func(_ *pruner.Service) {}),
	)

	switch tp {
	case node.Light:
		// LNs enforce pruning by default
		return fx.Module("prune",
			baseComponents,
			prunerService,
			// TODO(@walldiss @renaynay): remove conversion after Availability and Pruner interfaces are merged
			//  note this provide exists in pruner module to avoid cyclical imports
			fx.Provide(func(la *light.ShareAvailability) pruner.Pruner { return la }),
		)
	case node.Full:
		if cfg.EnableService {
			return fx.Module("prune",
				baseComponents,
				prunerService,
				fxutil.ProvideAs(full.NewPruner, new(pruner.Pruner)),
				fx.Supply([]fullavail.Option{}),
			)
		}
		return fx.Module("prune",
			baseComponents,
			fx.Invoke(func(ctx context.Context, ds datastore.Batching) error {
				return pruner.DetectPreviousRun(ctx, ds)
			}),
			fx.Supply([]fullavail.Option{fullavail.WithArchivalMode()}),
		)
	case node.Bridge:
		if cfg.EnableService {
			return fx.Module("prune",
				baseComponents,
				prunerService,
				fxutil.ProvideAs(full.NewPruner, new(pruner.Pruner)),
				fx.Supply([]fullavail.Option{}),
				fx.Supply([]core.Option{}),
			)
		}
		return fx.Module("prune",
			baseComponents,
			fx.Invoke(func(ctx context.Context, ds datastore.Batching) error {
				return pruner.DetectPreviousRun(ctx, ds)
			}),
			fx.Provide(func() []core.Option {
				return []core.Option{}
			}),
			fx.Supply([]fullavail.Option{fullavail.WithArchivalMode()}),
			fx.Supply([]core.Option{core.WithArchivalMode()}),
		)
	default:
		panic("unknown node type")
	}
}
