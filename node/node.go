package node

import (
	"context"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/routing"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/core"
)

var log = logging.Logger("node")

// Node represents the core structure of a Celestia node. It keeps references to all Celestia-specific
// components and services in one place and provides flexibility to run a Celestia node in different modes.
// Currently supported modes:
// * Full
// * Light
type Node struct {
	Type   Type
	Config *Config

	// the Node keeps a reference to the DI App that controls the lifecycles of services registered on the Node.
	app *fx.App

	// CoreClient provides access to Core Node
	CoreClient core.Client `optional:"true"`

	// p2p components
	ID          peer.ID
	Host        host.Host
	PeerStore   peerstore.Peerstore
	ConnManager connmgr.ConnManager
	ConnGater   connmgr.ConnectionGater
	Routing     routing.PeerRouting
	// p2p protocols
	PubSub *pubsub.PubSub
}

// newNode creates a new Node from given DI options.
// DI options allow initializing the Node with a customized set of components and services.
// NOTE: newNode is currently meant to be used privately to create various custom Node types e.g. full, unless we
// decide to give package users the ability to create custom node types themselves.
func newNode(opts ...fx.Option) (*Node, error) {
	node := new(Node)
	node.app = fx.New(
		fx.NopLogger,
		fx.Extract(node),
		fx.Options(opts...),
	)
	return node, node.app.Err()
}

// Start launches the Node and all the referenced components and services.
// Canceling the given context aborts the start.
func (n *Node) Start(ctx context.Context) error {
	log.Debugf("Starting %s Node...", n.Type)
	err := n.app.Start(ctx)
	if err != nil {
		log.Errorf("Error starting %s Node: %s", n.Type, err)
		return err
	}

	log.Infof("%s Node is started", n.Type)

	// TODO: Add bootstrapping
	return nil
}

// Stop shuts down the Node and all its running Components/Services.
// Canceling given context unblocks Stop and aborts graceful shutdown while forcing remaining Components/Services to
// close immediately.
func (n *Node) Stop(ctx context.Context) error {
	log.Debugf("Stopping %s Node...", n.Type)
	err := n.app.Stop(ctx)
	if err != nil {
		log.Errorf("Error stopping %s Node: %s", n.Type, err)
		return err
	}

	log.Infof("%s Node is stopped", n.Type)
	return nil
}
