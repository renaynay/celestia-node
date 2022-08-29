package p2p

import (
	"encoding/hex"
	"time"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/libs/fxutil"
)

// SetMutualPeers sets the `MutualPeers` field in the config.
func (cfg *Config) SetMutualPeers(addrs []string) {
	cfg.MutualPeers = addrs
}

// SetRefreshRoutingTablePeriod sets custom refresh period for dht.
// Currently, it is used to speed up tests.
func (cfg *Config) SetRefreshRoutingTablePeriod(interval time.Duration) {
	cfg.RoutingTableRefreshPeriod = interval
}

// WithP2PKey sets custom Ed25519 private key for p2p networking.
func WithP2PKey(key crypto.PrivKey) fx.Option {
	return fxutil.ReplaceAs(key, new(crypto.PrivKey))
}

// WithP2PKeyStr sets custom hex encoded Ed25519 private key for p2p networking.
func WithP2PKeyStr(key string) fx.Option {
	decKey, err := hex.DecodeString(key)
	if err != nil {
		return fx.Error(err)
	}

	privKey, err := crypto.UnmarshalEd25519PrivateKey(decKey)
	if err != nil {
		return fx.Error(err)
	}

	return fxutil.ReplaceAs(privKey, new(crypto.PrivKey))
}

// WithHost sets custom Host's data for p2p networking.
func WithHost(hst host.Host) fx.Option {
	return fxutil.ReplaceAs(hst, new(HostBase))
}
