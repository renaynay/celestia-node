package header

import (
	"github.com/celestiaorg/celestia-node/fraud"
	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/libs/fxutil"
	"github.com/celestiaorg/celestia-node/node/node"
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
)

var log = logging.Logger("header-module")

func Module(tp node.Type, cfg Config, options ...Option) fx.Option {
	sets := &settings{cfg: &cfg}
	for _, option := range options {
		option(sets)
	}
	baseOptions := fx.Options(
		fx.Supply(cfg),
		fx.Options(sets.opts...),
		fx.Provide(HeaderService),
		fx.Provide(HeaderStore),
		fx.Invoke(HeaderStoreInit(&cfg)),
		fxutil.ProvideAs(FraudService, new(fraud.Service), new(fraud.Subscriber)),
		fx.Provide(HeaderSyncer),
		fxutil.ProvideAs(P2PSubscriber, new(header.Broadcaster), new(header.Subscriber)),
		fx.Provide(HeaderP2PExchangeServer),
	)
	switch tp {
	case node.Light, node.Full:
		return fx.Module(
			"header",
			baseOptions,
			fx.Provide(HeaderExchangeP2P(cfg)),
		)
	case node.Bridge:
		return fx.Module(
			"header",
			baseOptions,
			fx.Supply(header.MakeExtendedHeader),
		)
	default:
		panic("wrong node type")
	}
}
