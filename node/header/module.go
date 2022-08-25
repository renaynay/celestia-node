package header

import (
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/fraud"
	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/libs/fxutil"
	"github.com/celestiaorg/celestia-node/node/node"
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
		fx.Provide(Service),
		fx.Provide(Store),
		fx.Invoke(InitStore(&cfg)),
		fxutil.ProvideAs(FraudService, new(fraud.Service), new(fraud.Subscriber)),
		fx.Provide(Syncer),
		fxutil.ProvideAs(P2PSubscriber, new(header.Broadcaster), new(header.Subscriber)),
		fx.Provide(P2PExchangeServer),
	)

	switch tp {
	case node.Light, node.Full:
		return fx.Module(
			"header",
			baseOptions,
			fx.Provide(P2PExchange(cfg)),
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
