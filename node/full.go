package node

import (
	"github.com/celestiaorg/celestia-node/node/block"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/node/p2p"
	"github.com/celestiaorg/celestia-node/node/rpc"
)

// NewFull assembles a new Full Node from required components.
func NewFull(cfg *Config) (*Node, error) {
	return newNode(fullComponents(cfg))
}

// fullComponents keeps all the components as DI options required to built a Full Node.
func fullComponents(cfg *Config) fx.Option {
	return fx.Options(
		// manual providing
		fx.Provide(func() Type {
			return Full
		}),
		fx.Provide(func() *Config {
			return cfg
		}),
		// components
		p2p.Components(cfg.P2P),
		rpc.Components(cfg.RPC),
		fx.Provide(block.Components),
	)
}
