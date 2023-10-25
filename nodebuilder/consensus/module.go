package consensus

import (
	"go.uber.org/fx"

	capp "github.com/celestiaorg/celestia-app/app"
	cnode "github.com/celestiaorg/celestia-app/node"

	"github.com/celestiaorg/celestia-node/core"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
)

func ConstructModule(tp node.Type, path string) fx.Option {
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
				func(publishFn capp.PublishFn) (*cnode.Node, error) {
					return newConsensusNode(path, publishFn)
				},
				fx.OnStart(func(nd *cnode.Node) error {
					return nd.Start()
				}),
				fx.OnStop(func(nd *cnode.Node) error {
					return nd.Stop()
				}),
			),
		),
		fx.Provide(func(nd *consensus.Node) core.Client {
			// TODO @cmwaters: return consensus node client
			return nil
		}),
	)
}

func newConsensusNode(path string, publishFn capp.PublishFn) (*cnode.Node, error) {
	fs, err := cnode.Load(path)
	if err != nil {
		return nil, err
	}

	return cnode.New(fs, publishFn)
}
