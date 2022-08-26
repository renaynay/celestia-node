package share

import (
	"github.com/celestiaorg/celestia-node/service/share"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	routingdisc "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"go.uber.org/fx"
)

// LightAvailability constructs light share availability.
func LightAvailability(cfg Config) func(
	bServ blockservice.BlockService,
	r routing.ContentRouting,
	h host.Host,
) *share.LightAvailability {
	return func(
		bServ blockservice.BlockService,
		r routing.ContentRouting,
		h host.Host,
	) *share.LightAvailability {
		disc := share.NewDiscovery(
			h,
			routingdisc.NewRoutingDiscovery(r),
			cfg.PeersLimit,
			cfg.DiscoveryInterval,
			cfg.AdvertiseInterval,
		)
		return share.NewLightAvailability(bServ, disc)
	}
}

// FullAvailability constructs full share availability.
func FullAvailability(cfg Config) func(
	bServ blockservice.BlockService,
	r routing.ContentRouting,
	h host.Host,
) *share.FullAvailability {
	return func(
		bServ blockservice.BlockService,
		r routing.ContentRouting,
		h host.Host,
	) *share.FullAvailability {
		disc := share.NewDiscovery(
			h,
			routingdisc.NewRoutingDiscovery(r),
			cfg.PeersLimit,
			cfg.DiscoveryInterval,
			cfg.AdvertiseInterval,
		)
		return share.NewFullAvailability(bServ, disc)
	}
}

// CacheAvailability wraps either Full or Light availability with a cache for result sampling.
func CacheAvailability[A share.Availability](lc fx.Lifecycle, ds datastore.Batching, avail A) share.Availability {
	ca := share.NewCacheAvailability(avail, ds)
	lc.Append(fx.Hook{
		OnStop: ca.Close,
	})
	return ca
}
