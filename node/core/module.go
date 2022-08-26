package core

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/libs/fxutil"
	"github.com/celestiaorg/celestia-node/node/node"

	"github.com/celestiaorg/celestia-node/core"
	"github.com/celestiaorg/celestia-node/header"
	headercore "github.com/celestiaorg/celestia-node/header/core"
)

// Module collects all the components and services related to managing the relationship with the Core node.
func Module(tp node.Type, cfg *Config, options ...Option) fx.Option {
	sets := &settings{cfg: cfg}
	for _, option := range options {
		option(sets)
	}
	switch tp {
	case node.Light, node.Full:
		return fx.Module("core", fx.Supply(*cfg), fx.Options(sets.opts...))
	case node.Bridge:
		return fx.Module("core",
			fx.Supply(*cfg),
			fx.Options(sets.opts...),
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
			fx.Provide(func(lc fx.Lifecycle) (core.Client, error) {
				if cfg.IP == "" {
					return nil, fmt.Errorf("no celestia-core endpoint given")
				}
				client, err := core.NewRemote(cfg.IP, cfg.RPCPort)
				if err != nil {
					return nil, err
				}
				lc.Append(fx.Hook{
					OnStart: func(context.Context) error {
						return client.Start()
					},
					OnStop: func(context.Context) error {
						return client.Stop()
					},
				})

				return client, err
			}),
		)
	default:
		panic("invalid node type")
	}
}
