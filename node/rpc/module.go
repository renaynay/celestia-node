package rpc

import (
	headerServ "github.com/celestiaorg/celestia-node/service/header"
	rpcServ "github.com/celestiaorg/celestia-node/service/rpc"
	stateServ "github.com/celestiaorg/celestia-node/service/state"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/node/node"
	shareServ "github.com/celestiaorg/celestia-node/service/share"
)

func Module(tp node.Type, cfg Config, options ...Option) fx.Option {
	sets := &settings{cfg: &cfg}
	for _, option := range options {
		option(sets)
	}
	switch tp {
	case node.Light, node.Full:
		return fx.Module(
			"rpc",
			fx.Provide(Server),
			fx.Invoke(Handler),
		)
	case node.Bridge:
		return fx.Module(
			"rpc",
			fx.Provide(Server),
			fx.Invoke(func(
				state *stateServ.Service,
				share *shareServ.Service,
				header *headerServ.Service,
				rpcSrv *rpcServ.Server,
			) {
				Handler(state, share, header, rpcSrv, nil)
			}),
		)
	default:
		panic("wrong node type")
	}
}
