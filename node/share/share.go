package share

import (
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	routingdisc "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/service/share"
)

// ShareService constructs new share.Service.
func ShareService(lc fx.Lifecycle, bServ blockservice.BlockService, avail share.Availability) *share.Service {
	service := share.NewService(bServ, avail)
	lc.Append(fx.Hook{
		OnStart: service.Start,
		OnStop:  service.Stop,
	})
	return service
}

// LightAvailability constructs light share availability.
func LightAvailability(cfg Config) func(
	lc fx.Lifecycle,
	bServ blockservice.BlockService,
	r routing.ContentRouting,
	h host.Host,
) *share.LightAvailability {
	return func(
		lc fx.Lifecycle,
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
		la := share.NewLightAvailability(bServ, disc)
		lc.Append(fx.Hook{
			OnStart: la.Start,
			OnStop:  la.Stop,
		})
		return la
	}
}

// FullAvailability constructs full share availability.
func FullAvailability(cfg Config) func(
	lc fx.Lifecycle,
	bServ blockservice.BlockService,
	r routing.ContentRouting,
	h host.Host,
) *share.FullAvailability {
	return func(
		lc fx.Lifecycle,
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
		fa := share.NewFullAvailability(bServ, disc)
		lc.Append(fx.Hook{
			OnStart: fa.Start,
			OnStop:  fa.Stop,
		})
		return fa
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
