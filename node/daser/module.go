package daser

import (
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/node/node"
)

func Module(tp node.Type) fx.Option {
	switch tp {
	case node.Light, node.Full:
		return fx.Module(
			"daser",
			fx.Provide(DASer),
		)
	case node.Bridge:
		return fx.Module("daser")
	default:
		panic("wrong node type")
	}
}
