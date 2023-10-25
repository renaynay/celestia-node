package relay

import (
	"go.uber.org/fx"

	capp "github.com/celestiaorg/celestia-app/app"

	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/relay"
)

func ConstructModule(tp node.Type) fx.Option {
	switch tp {
	case node.Light, node.Full, node.Bridge:
		return fx.Options()
	case node.Consensus:
	default:
		panic("invalid node type")
	}

	return fx.Module(
		"da_broadcaster",
		fx.Provide(relay.NewDARelayer),
		fx.Provide(func(relayer *relay.DARelayer) capp.PublishFn {
			return relayer.BroadcastAndStore
		}),
	)
}
