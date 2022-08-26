package p2p

import (
	"encoding/hex"
	"time"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/libs/fxutil"
)

// settings store values that can be augmented or changed for Node with Options.
type settings struct {
	cfg  *Config
	opts []fx.Option
}

type Option func(*settings)

// WithMutualPeers sets the `MutualPeers` field in the config.
func WithMutualPeers(addrs []string) Option {
	return func(sets *settings) {
		sets.cfg.MutualPeers = addrs
	}
}

// WithRefreshRoutingTablePeriod sets custom refresh period for dht.
// Currently, it is used to speed up tests.
func WithRefreshRoutingTablePeriod(interval time.Duration) Option {
	return func(sets *settings) {
		sets.cfg.RoutingTableRefreshPeriod = interval
	}
}

// WithP2PKey sets custom Ed25519 private key for p2p networking.
func WithP2PKey(key crypto.PrivKey) Option {
	return func(sets *settings) {
		sets.opts = append(sets.opts, fxutil.ReplaceAs(key, new(crypto.PrivKey)))
	}
}

// WithP2PKeyStr sets custom hex encoded Ed25519 private key for p2p networking.
func WithP2PKeyStr(key string) Option {
	return func(sets *settings) {
		decKey, err := hex.DecodeString(key)
		if err != nil {
			sets.opts = append(sets.opts, fx.Error(err))
			return
		}

		key, err := crypto.UnmarshalEd25519PrivateKey(decKey)
		if err != nil {
			sets.opts = append(sets.opts, fx.Error(err))
			return
		}

		sets.opts = append(sets.opts, fxutil.ReplaceAs(key, new(crypto.PrivKey)))
	}
}

// WithHost sets custom Host's data for p2p networking.
func WithHost(hst host.Host) Option {
	return func(sets *settings) {
		sets.opts = append(sets.opts, fxutil.ReplaceAs(hst, new(HostBase)))
	}
}
