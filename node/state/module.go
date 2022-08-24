package state

import (
	"context"

	"github.com/celestiaorg/celestia-node/node/header"

	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/node/node"

	"github.com/celestiaorg/celestia-node/fraud"
	"github.com/celestiaorg/celestia-node/libs/fxutil"
	"github.com/celestiaorg/celestia-node/service/state"
)

var log = logging.Logger("state-module")

// Module provides all components necessary to construct the
// state service.
func Module(tp node.Type, cfg Config, options ...Option) fx.Option {
	sets := &settings{cfg: &cfg}
	for _, option := range options {
		option(sets)
	}
	switch tp {
	case node.Light, node.Full, node.Bridge:
		return fx.Module(
			"state",
			fx.Supply(cfg),
			fx.Options(sets.opts...),
			fx.Provide(Keyring(sets.cfg)),
			fx.Provide(fx.Annotate(CoreAccessor,
				fx.OnStart(func(ctx context.Context, accessor state.Accessor) error {
					return accessor.Start(ctx)
				}),
				fx.OnStop(func(ctx context.Context, accessor state.Accessor) error {
					return accessor.Stop(ctx)
				}),
			)),
			fx.Provide(fx.Annotate(state.NewService,
				fx.OnStart(func(ctx context.Context, lc fx.Lifecycle, fservice fraud.Service, serv *state.Service) error {
					lifecycleCtx := fxutil.WithLifecycle(ctx, lc)
					return header.FraudLifecycle(ctx, lifecycleCtx, fraud.BadEncoding, fservice, serv.Start, serv.Stop)
				}),
				fx.OnStop(func(ctx context.Context, serv *state.Service) error {
					return serv.Stop(ctx)
				}),
			)),
		)
	default:
		panic("wrong node type")
	}
}
