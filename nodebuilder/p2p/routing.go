package p2p

import (
	"context"
	"fmt"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/libs/fxutil"
	nparams "github.com/celestiaorg/celestia-node/params"
)

var log = logging.Logger("node/p2p")

// ContentRouting constructs nil content routing,
// as for our use-case existing ContentRouting mechanisms, e.g DHT, are unsuitable
func ContentRouting(r routing.PeerRouting) routing.ContentRouting {
	return r.(*dht.IpfsDHT)
}

// PeerRouting provides constructor for PeerRouting over DHT.
// Basically, this provides a way to discover peer addresses by respecting public keys.
func PeerRouting(cfg Config) func(routingParams) (routing.PeerRouting, error) {
	return func(params routingParams) (routing.PeerRouting, error) {
		if cfg.RoutingTableRefreshPeriod <= 0 {
			cfg.RoutingTableRefreshPeriod = defaultRoutingRefreshPeriod
			log.Warnf("routingTableRefreshPeriod is not valid. restoring to default value: %d", cfg.RoutingTableRefreshPeriod)
		}
		opts := []dht.Option{
			dht.Mode(dht.ModeAuto),
			dht.BootstrapPeers(params.Peers...),
			dht.ProtocolPrefix(protocol.ID(fmt.Sprintf("/celestia/%s", params.Net))),
			dht.Datastore(params.DataStore),
			dht.RoutingTableRefreshPeriod(cfg.RoutingTableRefreshPeriod),
		}

		if cfg.Bootstrapper {
			// override options for bootstrapper
			opts = append(opts,
				dht.Mode(dht.ModeServer), // it must accept incoming connections
				dht.BootstrapPeers(),     // no bootstrappers for a bootstrapper ¯\_(ツ)_/¯
			)
		}

		d, err := dht.New(fxutil.WithLifecycle(params.Ctx, params.Lc), params.Host, opts...)
		if err != nil {
			return nil, err
		}
		params.Lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return d.Bootstrap(ctx)
			},
			OnStop: func(context.Context) error {
				return d.Close()
			},
		})
		return d, nil
	}
}

type routingParams struct {
	fx.In

	Ctx       context.Context
	Net       nparams.Network
	Peers     nparams.Bootstrappers
	Lc        fx.Lifecycle
	Host      HostBase
	DataStore datastore.Batching
}
