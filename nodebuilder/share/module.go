package share

import (
	"context"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	routingdisc "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/net/conngater"
	"go.uber.org/fx"

	libhead "github.com/celestiaorg/go-header"
	"github.com/celestiaorg/go-header/sync"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	modp2p "github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	modprune "github.com/celestiaorg/celestia-node/nodebuilder/pruner"
	"github.com/celestiaorg/celestia-node/pruner"
	"github.com/celestiaorg/celestia-node/share"
	"github.com/celestiaorg/celestia-node/share/availability/full"
	"github.com/celestiaorg/celestia-node/share/availability/light"
	"github.com/celestiaorg/celestia-node/share/eds"
	"github.com/celestiaorg/celestia-node/share/getters"
	disc "github.com/celestiaorg/celestia-node/share/p2p/discovery"
	"github.com/celestiaorg/celestia-node/share/p2p/peers"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexeds"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexnd"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexsub"
)

func ConstructModule(tp node.Type, cfg *Config, options ...fx.Option) fx.Option {
	// sanitize config values before constructing module
	cfgErr := cfg.Validate(tp)

	baseComponents := fx.Options(
		fx.Supply(*cfg),
		fx.Error(cfgErr),
		fx.Options(options...),
		fx.Provide(newShareModule),
		peerManagerComponents(tp, cfg),
		fullDiscoveryComponents(cfg),
		shrexSubComponents(),
		shrexGetterComponents(cfg),
		archivalComponents(cfg),
	)

	bridgeAndFullComponents := fx.Options(
		fx.Provide(getters.NewStoreGetter),
		shrexServerComponents(cfg),
		edsStoreComponents(cfg),
		fullAvailabilityComponents(),
		advertiseHooks(),
		fx.Provide(func(shrexSub *shrexsub.PubSub) shrexsub.BroadcastFn {
			return shrexSub.Broadcast
		}),
	)

	switch tp {
	case node.Bridge:
		return fx.Module(
			"share",
			baseComponents,
			bridgeAndFullComponents,
			fx.Provide(func() peers.Parameters {
				return cfg.PeerManagerParams
			}),
			fx.Provide(bridgeGetter),
			fx.Invoke(func(lc fx.Lifecycle, sub *shrexsub.PubSub) error {
				lc.Append(fx.Hook{
					OnStart: sub.Start,
					OnStop:  sub.Stop,
				})
				return nil
			}),
		)
	case node.Full:
		return fx.Module(
			"share",
			baseComponents,
			bridgeAndFullComponents,
			fx.Provide(getters.NewIPLDGetter),
			fx.Provide(fullGetter),
		)
	case node.Light:
		return fx.Module(
			"share",
			baseComponents,
			lightAvailabilityComponents(cfg),
			fx.Invoke(ensureEmptyEDSInBS),
			fx.Provide(getters.NewIPLDGetter),
			fx.Provide(lightGetter),
			// shrexsub broadcaster stub for daser
			fx.Provide(func() shrexsub.BroadcastFn {
				return func(context.Context, shrexsub.Notification) error {
					return nil
				}
			}),
			// needed to invoke archival discovery lifecycle as no other
			// component takes it.
			fx.Invoke(func(_ map[string]*disc.Discovery) {}),
		)
	default:
		panic("invalid node type")
	}
}

func fullDiscoveryComponents(cfg *Config) fx.Option {
	return fx.Options(
		fx.Invoke(func(_ *disc.Discovery) {}),
		fx.Provide(fx.Annotate(
			newFullDiscovery(cfg.Discovery),
			fx.OnStart(func(ctx context.Context, d *disc.Discovery) error {
				return d.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, d *disc.Discovery) error {
				return d.Stop(ctx)
			}),
		)),
	)
}

func peerManagerComponents(tp node.Type, cfg *Config) fx.Option {
	switch tp {
	case node.Full, node.Light:
		return fx.Options(
			fx.Provide(func() peers.Parameters {
				return cfg.PeerManagerParams
			}),
			fx.Provide(
				func(
					params peers.Parameters,
					host host.Host,
					connGater *conngater.BasicConnectionGater,
					shrexSub *shrexsub.PubSub,
					headerSub libhead.Subscriber[*header.ExtendedHeader],
					// we must ensure Syncer is started before PeerManager
					// so that Syncer registers header validator before PeerManager subscribes to headers
					_ *sync.Syncer[*header.ExtendedHeader],
				) (*peers.Manager, error) {
					return peers.NewManager(
						params,
						host,
						connGater,
						peers.WithShrexSubPools(shrexSub, headerSub),
						peers.WithTag("full"),
					)
				},
			),
		)
	case node.Bridge:
		return fx.Provide(func(
			params peers.Parameters,
			host host.Host,
			connGater *conngater.BasicConnectionGater,
		) (*peers.Manager, error) {
			return peers.NewManager(params, host, connGater, peers.WithTag("full"))
		})
	default:
		panic("invalid node type")
	}
}

func shrexSubComponents() fx.Option {
	return fx.Provide(
		func(ctx context.Context, h host.Host, network modp2p.Network) (*shrexsub.PubSub, error) {
			return shrexsub.NewPubSub(ctx, h, network.String())
		},
	)
}

// shrexGetterComponents provides components for a shrex getter that
// is capable of requesting
func shrexGetterComponents(cfg *Config) fx.Option {
	return fx.Options(
		// shrex-nd client
		fx.Provide(
			func(host host.Host, network modp2p.Network) (*shrexnd.Client, error) {
				cfg.ShrExNDParams.WithNetworkID(network.String())
				return shrexnd.NewClient(cfg.ShrExNDParams, host)
			},
		),

		// shrex-eds client
		fx.Provide(
			func(host host.Host, network modp2p.Network) (*shrexeds.Client, error) {
				cfg.ShrExEDSParams.WithNetworkID(network.String())
				return shrexeds.NewClient(cfg.ShrExEDSParams, host)
			},
		),

		// shrex-getter
		fx.Provide(fx.Annotate(
			func(
				edsClient *shrexeds.Client,
				ndClient *shrexnd.Client,
				man *peers.Manager,
				window pruner.AvailabilityWindow,
				opts []getters.Option,
			) *getters.ShrexGetter {
				opts = append(opts, getters.WithAvailabilityWindow(window))
				return getters.NewShrexGetter(edsClient, ndClient, man, opts...)
			},
			fx.OnStart(func(ctx context.Context, getter *getters.ShrexGetter) error {
				return getter.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, getter *getters.ShrexGetter) error {
				return getter.Stop(ctx)
			}),
		)),
	)
}

func shrexServerComponents(cfg *Config) fx.Option {
	return fx.Options(
		fx.Invoke(func(_ *shrexeds.Server, _ *shrexnd.Server) {}),
		fx.Provide(fx.Annotate(
			func(host host.Host, store *eds.Store, network modp2p.Network) (*shrexeds.Server, error) {
				cfg.ShrExEDSParams.WithNetworkID(network.String())
				return shrexeds.NewServer(cfg.ShrExEDSParams, host, store)
			},
			fx.OnStart(func(ctx context.Context, server *shrexeds.Server) error {
				return server.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, server *shrexeds.Server) error {
				return server.Stop(ctx)
			}),
		)),
		fx.Provide(fx.Annotate(
			func(
				host host.Host,
				store *eds.Store,
				network modp2p.Network,
			) (*shrexnd.Server, error) {
				cfg.ShrExNDParams.WithNetworkID(network.String())
				return shrexnd.NewServer(cfg.ShrExNDParams, host, store)
			},
			fx.OnStart(func(ctx context.Context, server *shrexnd.Server) error {
				return server.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, server *shrexnd.Server) error {
				return server.Stop(ctx)
			})),
		),
	)
}

func edsStoreComponents(cfg *Config) fx.Option {
	return fx.Options(
		fx.Provide(fx.Annotate(
			func(path node.StorePath, ds datastore.Batching) (*eds.Store, error) {
				return eds.NewStore(cfg.EDSStoreParams, string(path), ds)
			},
			fx.OnStart(func(ctx context.Context, store *eds.Store) error {
				err := store.Start(ctx)
				if err != nil {
					return err
				}
				return ensureEmptyCARExists(ctx, store)
			}),
			fx.OnStop(func(ctx context.Context, store *eds.Store) error {
				return store.Stop(ctx)
			}),
		)),
	)
}

func fullAvailabilityComponents() fx.Option {
	return fx.Options(
		fx.Provide(full.NewShareAvailability),
		fx.Provide(func(avail *full.ShareAvailability) share.Availability {
			return avail
		}),
	)
}

func lightAvailabilityComponents(cfg *Config) fx.Option {
	return fx.Options(
		fx.Provide(fx.Annotate(
			light.NewShareAvailability,
			fx.OnStop(func(ctx context.Context, la *light.ShareAvailability) error {
				return la.Close(ctx)
			}),
		)),
		fx.Provide(func() []light.Option {
			return []light.Option{
				light.WithSampleAmount(cfg.LightAvailability.SampleAmount),
			}
		}),
		fx.Provide(func(avail *light.ShareAvailability) share.Availability {
			return avail
		}),
	)
}

func archivalComponents(cfg *Config) fx.Option {
	return fx.Options(
		archivalDiscovery(cfg),
		archivalPeerManager(),
	)
}

func archivalDiscovery(cfg *Config) fx.Option {
	return fx.Options(
		fx.Provide(func(
			lc fx.Lifecycle,
			tp node.Type,
			pruneCfg *modprune.Config,
			d *disc.Discovery,
			h host.Host,
			r routing.ContentRouting,
			opt disc.Option,
		) (map[string]*disc.Discovery, error) {
			// if pruner is enabled for BN, no archival service is necessary
			if tp == node.Bridge && pruneCfg.EnableService {
				return map[string]*disc.Discovery{"full": d}, nil
			}

			archivalDisc, err := disc.NewDiscovery(
				cfg.Discovery,
				h,
				routingdisc.NewRoutingDiscovery(r),
				archivalNodesTag,
				opt,
			)
			if err != nil {
				return nil, err
			}

			lc.Append(fx.Hook{
				OnStart: archivalDisc.Start,
				OnStop:  archivalDisc.Stop,
			})

			return map[string]*disc.Discovery{"full": d, "archival": archivalDisc}, nil
		}),
	)
}

func archivalPeerManager() fx.Option {
	return fx.Provide(func(
		lc fx.Lifecycle,
		params peers.Parameters,
		h host.Host,
		gater *conngater.BasicConnectionGater,
		fullManager *peers.Manager,
	) ([]*peers.Manager, []getters.Option, disc.Option, error) {
		archivalPeerManager, err := peers.NewManager(params, h, gater, peers.WithTag("archival"))
		if err != nil {
			return nil, nil, nil, err
		}
		lc.Append(fx.Hook{
			OnStart: archivalPeerManager.Start,
			OnStop:  archivalPeerManager.Stop,
		})

		managers := []*peers.Manager{fullManager, archivalPeerManager}
		opts := []getters.Option{getters.WithArchivalPeerManager(archivalPeerManager)}
		return managers, opts, disc.WithOnPeersUpdate(archivalPeerManager.UpdateNodePool), nil
	})
}

// advertiseHooks provides the lifecycle hooks for both `full` and, if
// applicable, `archival` discovery.
func advertiseHooks() fx.Option {
	return fx.Invoke(func(lc fx.Lifecycle, tp node.Type, cfg *modprune.Config, discs map[string]*disc.Discovery) error {
		if tp == node.Light {
			// do not advertise on `full` or `archival` topics
			return nil
		}

		// start advertising on `full` node topic as both FNs and BNs (regardless of pruning status)
		// are considered `full` on the celestia DA network
		fullDisc, ok := discs["full"]
		if !ok {
			return fmt.Errorf("expected full discovery component to be provided")
		}
		lc.Append(fx.Hook{
			OnStart: fullDisc.StartAdvertising,
			OnStop:  fullDisc.StopAdvertising,
		})

		// if the BN or FN is not in pruning mode, it is `archival`, so
		// advertise on the `archival` topic
		if !cfg.EnableService {
			archivalDisc, ok := discs["archival"]
			if !ok {
				return fmt.Errorf("expected archival discovery component to be provided")
			}
			lc.Append(fx.Hook{
				OnStart: archivalDisc.StartAdvertising,
				OnStop:  archivalDisc.StopAdvertising,
			})
		}

		return nil
	})
}
