package p2p

import (
	"fmt"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/fx"
)

// Config combines all configuration fields for P2P subsystem.
type Config struct {
	// ListenAddresses - Addresses to listen to on local NIC.
	ListenAddresses []string
	// AnnounceAddresses - Addresses to be announced/advertised for peers to connect to
	AnnounceAddresses []string
	// NoAnnounceAddresses - Addresses the P2P subsystem may know about, but that should not be announced/advertised,
	// as undialable from WAN
	NoAnnounceAddresses []string
	// TODO(@Wondertan): This should be a built-time parameter. See https://github.com/celestiaorg/celestia-node/issues/62
	// Networks stands for network name, e.g. celestia-devnet.
	Network string
	// TODO(@Wondertan): This should be a built-time parameter. See https://github.com/celestiaorg/celestia-node/issues/63
	// Bootstrapper is flag telling this node is a bootstrapper.
	Bootstrapper bool
	// BootstrapPeers is a list of network specific peers that help with network bootstrapping.
	BootstrapPeers []string
	// MutualPeers are peers which have a bidirectional peering agreement with the configured node.
	// Connections with those peers are protected from being trimmed, dropped or negatively scored.
	// NOTE: Any two peers must bidirectionally configure each other on their MutualPeers field.
	MutualPeers []string
	// PeerExchange configures the node, whether it should share some peers to a pruned peer.
	// This is enabled by default for Bootstrappers.
	PeerExchange bool
	// ConnManager is a configuration tuple for ConnectionManager.
	ConnManager ConnManagerConfig
}

// DefaultConfig returns default configuration for P2P subsystem.
func DefaultConfig() Config {
	return Config{
		ListenAddresses: []string{
			"/ip4/0.0.0.0/tcp/2121",
			"/ip6/::/tcp/2121",
		},
		AnnounceAddresses: []string{},
		NoAnnounceAddresses: []string{
			"/ip4/0.0.0.0/tcp/2121",
			"/ip4/127.0.0.1/tcp/2121",
			"/ip6/::/tcp/2121",
		},
		Network:        "devnet",
		BootstrapPeers: []string{},
		MutualPeers:    []string{},
		Bootstrapper:   false,
		PeerExchange:   false,
		ConnManager:    DefaultConnManagerConfig(),
	}
}

// Components collects all the components and services related to p2p.
func Components(cfg Config) fx.Option {
	return fx.Options(
		fx.Provide(Identity),
		fx.Provide(PeerStore),
		fx.Provide(ConnectionManager(cfg)),
		fx.Provide(ConnectionGater),
		fx.Provide(Host(cfg)),
		fx.Provide(RoutedHost),
		fx.Provide(PubSub(cfg)),
		fx.Provide(DataExchange(cfg)),
		fx.Provide(DAG),
		fx.Provide(PeerRouting(cfg)),
		fx.Provide(ContentRouting),
		fx.Provide(AddrsFactory(cfg.AnnounceAddresses, cfg.NoAnnounceAddresses)),
		fx.Invoke(Listen(cfg.ListenAddresses)),
	)
}

func (cfg *Config) bootstrapPeers() (_ []peer.AddrInfo, err error) {
	fmt.Println("HELLO I SAID HELLOOOOOOOOOOOOOO:    ", cfg.BootstrapPeers[0])

	maddrs := make([]ma.Multiaddr, len(cfg.BootstrapPeers))
	for i, addr := range cfg.BootstrapPeers {
		maddrs[i], err = ma.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("failure to parse config.P2P.BootstrapPeers: %s", err)
		}
	}

	return peer.AddrInfosFromP2pAddrs(maddrs...)
}

func (cfg *Config) mutualPeers() (_ []peer.AddrInfo, err error) {
	maddrs := make([]ma.Multiaddr, len(cfg.MutualPeers))
	for i, addr := range cfg.MutualPeers {
		maddrs[i], err = ma.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("failure to parse config.P2P.MutualPeers: %s", err)
		}
	}

	return peer.AddrInfosFromP2pAddrs(maddrs...)
}
