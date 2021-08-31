package node

import (
	"context"

	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/node/config"
	"github.com/celestiaorg/celestia-node/node/p2p"
	"github.com/celestiaorg/celestia-node/rpc"
)

// NewFull creates and runs a new ready-to-go Full Node.
// To gracefully stop it the Stop method must be used.
func NewFull(ctx context.Context, cfg *config.Config) (*Node, error) {
	node, err := newNode(cfg, full(cfg))
	if err != nil {
		return nil, err
	}

	err = node.Start(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: Add bootstrapping

	node.tp = Full
	return node, nil
}

// full keeps all the DI options required to built a Full Node.
func full(cfg *config.Config) fx.Option {
	return fx.Options(
		fx.Provide(p2p.Host),
		fx.Provide(RPCClientConstructor(cfg)),
	)
}

func RPCClientConstructor(cfg *config.Config) interface{} {
	return func() (*rpc.Client, error) {
		return rpc.NewClient(cfg.RPCConfig.Protocol, cfg.RPCConfig.RemoteAddr)
	}
}
