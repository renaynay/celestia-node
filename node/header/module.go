package header

import (
	"context"
	"github.com/celestiaorg/celestia-node/header/p2p"
	"github.com/celestiaorg/celestia-node/header/store"
	"github.com/celestiaorg/celestia-node/header/sync"
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/fraud"
	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/libs/fxutil"
	"github.com/celestiaorg/celestia-node/node/node"
)

var log = logging.Logger("header-module")

func Module(tp node.Type, cfg *Config, options ...Option) fx.Option {
	sets := &settings{cfg: cfg}
	for _, option := range options {
		option(sets)
	}
	baseOptions := fx.Options(
		fx.Supply(cfg),
		fx.Options(sets.opts...),
		fx.Provide(Service),
		fx.Provide(fx.Annotate(
			store.NewStore,
			fx.OnStart(func(ctx context.Context, store header.Store) error {
				return store.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, store header.Store) error {
				return store.Stop(ctx)
			}),
		)),
		fx.Invoke(InitStore(cfg)),
		fxutil.ProvideAs(FraudService, new(fraud.Service), new(fraud.Subscriber)),
		fx.Provide(func(subscriber *p2p.Subscriber) header.Subscriber {
			return subscriber
		}),
		fx.Provide(func(subscriber *p2p.Subscriber) header.Broadcaster {
			return subscriber
		}),
		fx.Provide(fx.Annotate(
			sync.NewSyncer,
			fx.OnStart(func(ctx context.Context, lc fx.Lifecycle, fservice fraud.Service, syncer *sync.Syncer) error {
				lifecycleCtx := fxutil.WithLifecycle(ctx, lc)
				return FraudLifecycle(ctx, lifecycleCtx, fraud.BadEncoding, fservice, syncer.Start, syncer.Stop)
			}),
			fx.OnStop(func(ctx context.Context, syncer *sync.Syncer) error {
				return syncer.Stop(ctx)
			}),
		)),
		fx.Provide(fx.Annotate(
			p2p.NewSubscriber,
			fx.OnStart(func(ctx context.Context, sub *p2p.Subscriber) error {
				return sub.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, sub *p2p.Subscriber) error {
				return sub.Stop(ctx)
			}),
		)),

		fx.Provide(fx.Annotate(
			p2p.NewExchangeServer,
			fx.OnStart(func(ctx context.Context, server *p2p.ExchangeServer) error {
				return server.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, server *p2p.ExchangeServer) error {
				return server.Stop(ctx)
			}),
		)),
	)

	switch tp {
	case node.Light, node.Full:
		return fx.Module(
			"header",
			baseOptions,
			fx.Provide(P2PExchange(*cfg)),
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
