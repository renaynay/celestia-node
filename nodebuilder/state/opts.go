package state

import (
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-app/x/payment/types"
)

type Option func(*settings)

// settings store values that can be augmented or changed for Node with Options.
type settings struct {
	cfg  *Config
	opts []fx.Option
}

// WithKeyringAccName sets the `KeyringAccName` field in the key config.
func WithKeyringAccName(name string) Option {
	return func(sets *settings) {
		sets.cfg.KeyringAccName = name
	}
}

// WithKeyringSigner overrides the default keyring signer constructed
// by the node.
func WithKeyringSigner(signer *types.KeyringSigner) Option {
	return func(sets *settings) {
		sets.opts = append(sets.opts, fx.Replace(signer))
	}
}
