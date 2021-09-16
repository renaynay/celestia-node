package p2p

import (
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/fx"
)

// TODO(Wondertan): Some of the fields should not be configurable,
//  but rather be built-in into the binary, like Network and Bootstrapper
// Config combines all configuration fields for P2P subsystem.
type Config struct {
	// ListenAddresses - Addresses to listen to on local NIC.
	ListenAddresses []string
	// AnnounceAddresses - Addresses to be announced/advertised for peers to connect to
	AnnounceAddresses []string
	// NoAnnounceAddresses - Addresses the P2P subsystem may know about, but that should not be announced/advertised,
	// as undialable from WAN
	NoAnnounceAddresses []string

	Network        string
	Bootstrapper   bool
	BootstrapPeers []string
	FriendPeers    []string
	PeerExchange   bool
	ConnManager    *ConnManagerConfig
}

// DefaultConfig returns default configuration for P2P subsystem.
func DefaultConfig() *Config {
	return &Config{
		Network: "devnet",
		ListenAddresses: []string{
			"/ip4/0.0.0.0/tcp/2121",
			"/ip6/::/tcp/2121",
		},
		NoAnnounceAddresses: []string{
			"/ip4/0.0.0.0/tcp/2121",
			"/ip4/127.0.0.1/tcp/2121",
			"/ip6/::/tcp/2121",
		},
		BootstrapPeers: nil,
		Bootstrapper:   false,
		PeerExchange:   false,
		ConnManager:    DefaultConnManagerConfig(),
	}
}

// Components collects all the components and services related to p2p.
func Components(cfg *Config) fx.Option {
	return fx.Options(
		// TODO(Wondertan): This shouldn't be here, but it is required until we start using real datastore
		fx.Provide(func() datastore.Batching {
			return datastore.NewMapDatastore()
		}),

		fx.Provide(Identity),
		fx.Provide(PeerStore),
		fx.Provide(ConnectionManager(cfg)),
		fx.Provide(ConnectionGater),
		fx.Provide(Host(cfg)),
		fx.Provide(RoutedHost),
		fx.Provide(PubSub(cfg)),
		fx.Provide(Routing(cfg)),
		fx.Provide(AddrsFactory(cfg.AnnounceAddresses, cfg.NoAnnounceAddresses)),
		fx.Invoke(Listen(cfg.ListenAddresses)),
	)
}

func (cfg *Config) bootstrapPeers() (_ []peer.AddrInfo, err error) {
	maddrs := make([]ma.Multiaddr, len(cfg.BootstrapPeers))
	for i, addr := range cfg.BootstrapPeers {
		maddrs[i], err = ma.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("failure to parse config.P2P.BootstrapPeers: %s", err)
		}
	}

	return peer.AddrInfosFromP2pAddrs(maddrs...)
}

func (cfg *Config) friendPeers() (_ []peer.AddrInfo, err error) {
	maddrs := make([]ma.Multiaddr, len(cfg.FriendPeers))
	for i, addr := range cfg.BootstrapPeers {
		maddrs[i], err = ma.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("failure to parse config.P2P.FriendPeers: %s", err)
		}
	}

	return peer.AddrInfosFromP2pAddrs(maddrs...)
}
