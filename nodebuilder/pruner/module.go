package pruner

import (
	"context"

	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/core"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/pruner"
	"github.com/celestiaorg/celestia-node/pruner/archival"
	"github.com/celestiaorg/celestia-node/pruner/full"
	"github.com/celestiaorg/celestia-node/pruner/light"
	"github.com/celestiaorg/celestia-node/share/eds"
)

func ConstructModule(tp node.Type, cfg *Config) fx.Option {
	baseComponents := fx.Options(
		fx.Supply(cfg),
	)

	if !cfg.EnableService {
		return disabledPrunerComponents(tp, baseComponents)
	}

	baseComponents = fx.Options(
		baseComponents,
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
	case node.Full:
		return fx.Module("prune",
			baseComponents,
			fx.Provide(func(store *eds.Store) pruner.Pruner {
				return full.NewPruner(store)
			}),
			fx.Provide(func() pruner.AvailabilityWindow {
				return full.Window
			}),
		)
	case node.Bridge:
		return fx.Module("prune",
			baseComponents,
			fx.Provide(func(store *eds.Store) pruner.Pruner {
				return full.NewPruner(store)
			}),
			fx.Provide(func() pruner.AvailabilityWindow {
				return full.Window
			}),
			fx.Provide(func(window pruner.AvailabilityWindow) []core.Option {
				return []core.Option{core.WithAvailabilityWindow(window)}
			}),
		)
	// TODO: Eventually, light nodes will be capable of pruning samples
	//  in which case, this can be enabled.
	case node.Light:
		return fx.Module("prune",
			fx.Provide(func() pruner.AvailabilityWindow {
				return light.Window
			}),
		)
	default:
		panic("unknown node type")
	}
}

func disabledPrunerComponents(tp node.Type, baseComponents fx.Option) fx.Option {
	switch tp {
	case node.Light:
		// light nodes are still subject to sampling within window
		// even if pruning is not enabled.
		return fx.Options(
			baseComponents,
			fx.Provide(func() pruner.AvailabilityWindow {
				return light.Window
			}),
		)
	case node.Full:
		return fx.Options(
			baseComponents,
			fx.Provide(func() pruner.AvailabilityWindow {
				return archival.Window
			}),
		)
	case node.Bridge:
		return fx.Options(
			baseComponents,
			fx.Provide(func() pruner.AvailabilityWindow {
				return archival.Window
			}),
			fx.Provide(func() []core.Option {
				return []core.Option{}
			}),
		)
	default:
		panic("unknown node type")
	}
}
