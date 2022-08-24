package share

import (
	"github.com/celestiaorg/celestia-node/node/node"
	"github.com/celestiaorg/celestia-node/service/share"
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
)

var log = logging.Logger("share-module")

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
			fx.Provide(ShareService),
			fx.Provide(LightAvailability(cfg)),
			fx.Provide(CacheAvailability[*share.LightAvailability]),
		)
	case node.Bridge, node.Full:
		return fx.Module(
			"share",
			fx.Supply(cfg),
			fx.Options(sets.opts...),
			fx.Invoke(share.EnsureEmptySquareExists),
			fx.Provide(ShareService),
			fx.Provide(FullAvailability(cfg)),
			fx.Provide(CacheAvailability[*share.FullAvailability]),
		)
	default:
		panic("wrong node type")
	}
}
