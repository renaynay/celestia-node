package share

import (
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/node/node"
	"github.com/celestiaorg/celestia-node/service/share"
)

func Module(tp node.Type, cfg Config, options ...Option) fx.Option {
	sets := &settings{cfg: &cfg}
	for _, option := range options {
		option(sets)
	}
	switch tp {
	case node.Light:
		return fx.Module(
			"share",
			fx.Supply(cfg),
			fx.Options(sets.opts...),
			fx.Invoke(share.EnsureEmptySquareExists),
			fx.Provide(Service),
			fx.Provide(LightAvailability(cfg)),
			fx.Provide(CacheAvailability[*share.LightAvailability]),
		)
	case node.Bridge, node.Full:
		return fx.Module(
			"share",
			fx.Supply(cfg),
			fx.Options(sets.opts...),
			fx.Invoke(share.EnsureEmptySquareExists),
			fx.Provide(Service),
			fx.Provide(FullAvailability(cfg)),
			fx.Provide(CacheAvailability[*share.FullAvailability]),
		)
	default:
		panic("wrong node type")
	}
}
