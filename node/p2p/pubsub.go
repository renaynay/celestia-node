package p2p

import (
	"context"

	"github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/node/fxutil"
)

func PubSub(cfg *Config) func(ctx context.Context, lc fx.Lifecycle, host host.Host) (*pubsub.PubSub, error) {
	return func(ctx context.Context, lc fx.Lifecycle, host host.Host) (*pubsub.PubSub, error) {
		fpeers, err := cfg.friendPeers()
		if err != nil {
			return nil, err
		}

		// TODO for PubSub options:
		//  * Hash-based MsgId function.
		//  * Validate default peer scoring params for our use-case.
		//  * Strict subscription filter
		//  * For different network types(mainnet/testnet/devnet) we should have different network topic names.
		//  * Hardcode positive score for bootstrap peers
		//  * Bootstrappers should only gossip and PX
		//  * Peers should trust boostrappers, so peerscore for them should always be high.
		opts := []pubsub.Option{
			pubsub.WithPeerExchange(cfg.PeerExchange || cfg.Bootstrapper),
			pubsub.WithDirectPeers(fpeers),
		}

		return pubsub.NewGossipSub(fxutil.WithLifecycle(ctx, lc), host, opts...)
	}
}
