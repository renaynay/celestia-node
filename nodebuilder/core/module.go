package core

import (
	"context"

	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/core"
	"github.com/celestiaorg/celestia-node/header"
	headercore "github.com/celestiaorg/celestia-node/header/core"
	"github.com/celestiaorg/celestia-node/libs/fxutil"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
)

// Module collects all the components and services related to managing the relationship with the Core node.
func Module(tp node.Type, cfg *Config, options ...fx.Option) fx.Option {
	switch tp {
	case node.Light, node.Full:
		return fx.Module("core", fx.Supply(*cfg), fx.Options(options...))
	case node.Bridge:
		return fx.Module("core",
			fx.Supply(*cfg),
			fx.Invoke(cfg.ValidateBasic),
			fx.Options(options...),
			fx.Provide(core.NewBlockFetcher),
			fxutil.ProvideAs(headercore.NewExchange, new(header.Exchange)),
			fx.Invoke(fx.Annotate(
				headercore.NewListener,
				fx.OnStart(func(ctx context.Context, listener *headercore.Listener) error {
					return listener.Start(ctx)
				}),
				fx.OnStop(func(ctx context.Context, listener *headercore.Listener) error {
					return listener.Stop(ctx)
				}),
			)),
			fx.Provide(fx.Annotate(
				Remote(*cfg),
				fx.OnStart(func(ctx context.Context, client core.Client) error {
					return client.Start()
				}),
				fx.OnStop(func(ctx context.Context, client core.Client) error {
					return client.Stop()
				}),
			)),
		)
	default:
		panic("invalid node type")
	}
}
