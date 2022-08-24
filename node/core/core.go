package core

import (
	"context"
	"fmt"

	"github.com/ipfs/go-blockservice"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/node/node"

	"github.com/celestiaorg/celestia-node/libs/fxutil"

	"github.com/celestiaorg/celestia-node/core"
	"github.com/celestiaorg/celestia-node/header"
	headercore "github.com/celestiaorg/celestia-node/header/core"
)

// Config combines all configuration fields for managing the relationship with a Core node.
type Config struct {
	IP       string
	RPCPort  string
	GRPCPort string
}

// DefaultConfig returns default configuration for managing the
// node's connection to a Celestia-Core endpoint.
func DefaultConfig() Config {
	return Config{}
}

// Module collects all the components and services related to managing the relationship with the Core node.
func Module(tp node.Type, cfg Config, options ...Option) fx.Option {
	sets := &settings{cfg: &cfg}
	for _, option := range options {
		option(sets)
	}
	switch tp {
	case node.Light, node.Full:
		return fx.Module("core", fx.Supply(cfg), fx.Options(sets.opts...))
	case node.Bridge:
		return fx.Module("core",
			fx.Supply(cfg),
			fx.Options(sets.opts...),
			fx.Provide(core.NewBlockFetcher),
			fxutil.ProvideAs(headercore.NewExchange, new(header.Exchange)),
			fx.Invoke(HeaderListener),
			fx.Provide(func(lc fx.Lifecycle) (core.Client, error) {
				if cfg.IP == "" {
					return nil, fmt.Errorf("no celestia-core endpoint given")
				}
				client, err := RemoteClient(cfg)
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

func HeaderListener(
	lc fx.Lifecycle,
	ex *core.BlockFetcher,
	bcast header.Broadcaster,
	bServ blockservice.BlockService,
	construct header.ConstructFn,
) *headercore.Listener {
	cl := headercore.NewListener(bcast, ex, bServ, construct)
	lc.Append(fx.Hook{
		OnStart: cl.Start,
		OnStop:  cl.Stop,
	})
	return cl
}

// RemoteClient provides a constructor for core.Client over RPC.
func RemoteClient(cfg Config) (core.Client, error) {
	return core.NewRemote(cfg.IP, cfg.RPCPort)
}
