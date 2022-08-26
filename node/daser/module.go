package daser

import (
	"context"
	"github.com/celestiaorg/celestia-node/das"
	"github.com/celestiaorg/celestia-node/fraud"
	"github.com/celestiaorg/celestia-node/libs/fxutil"
	header "github.com/celestiaorg/celestia-node/node/header"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/node/node"
)

func Module(tp node.Type) fx.Option {
	switch tp {
	case node.Light, node.Full:
		return fx.Module(
			"daser",
			fx.Provide(fx.Annotate(
				DASer,
				fx.OnStart(func(ctx context.Context, lc fx.Lifecycle, fservice fraud.Service, das *das.DASer) error {
					lifecycleCtx := fxutil.WithLifecycle(ctx, lc)
					return header.FraudLifecycle(ctx, lifecycleCtx, fraud.BadEncoding, fservice, das.Start, das.Stop)
				}),
				fx.OnStop(func(ctx context.Context, das *das.DASer) error {
					return das.Stop(ctx)
				}),
			)),
		)
	case node.Bridge:
		return fx.Module("daser")
	default:
		panic("wrong node type")
	}
}
