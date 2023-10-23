package consensus

import (
	"context"

	"go.uber.org/fx"

	consensus "github.com/celestiaorg/celestia-app/node"

	"github.com/celestiaorg/celestia-node/nodebuilder/node"
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
		"consensus",
		fx.Provide(
			fx.Annotate(
				newConsensusNode,
				fx.OnStart(func(
					ctx context.Context,
					nd *consensus.Node,
					publishFn consensus.PublishFn,
				) error {
					return nd.Run(ctx, publishFn)
				}),
				fx.OnStop(func(ctx context.Context, nd *consensus.Node) error {
					// TODO @cmwaters: stop consensus node
					return nil
				}),
			),
		),
	)
}

func newConsensusNode() (*consensus.Node, error) {
	// TODO @renaynay @cmwaters: needs to be an actual constructor somehow
	return &consensus.Node{}, nil
}
