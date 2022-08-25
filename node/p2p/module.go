package p2p

import (
	"go.uber.org/fx"
)

// Module collects all the components and services related to p2p.
func Module(cfg *Config, options ...Option) fx.Option {
	sets := &settings{cfg: cfg}
	for _, option := range options {
		option(sets)
	}

	return fx.Module(
		"p2p",
		fx.Supply(cfg),
		fx.Options(sets.opts...),
		fx.Provide(Key),
		fx.Provide(ID),
		fx.Provide(PeerStore),
		fx.Provide(ConnectionManager(*cfg)),
		fx.Provide(ConnectionGater),
		fx.Provide(Host(*cfg)),
		fx.Provide(RoutedHost),
		fx.Provide(PubSub(*cfg)),
		fx.Provide(DataExchange(*cfg)),
		fx.Provide(BlockService),
		fx.Provide(PeerRouting(*cfg)),
		fx.Provide(ContentRouting),
		fx.Provide(AddrsFactory(cfg.AnnounceAddresses, cfg.NoAnnounceAddresses)),
		fx.Invoke(Listen(cfg.ListenAddresses)),
	)
}
