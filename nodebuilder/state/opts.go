package state

import (
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-app/x/payment/types"
)

// SetKeyringAccName sets the `KeyringAccName` field in the key config.
func (cfg *Config) SetKeyringAccName(name string) {
	cfg.KeyringAccName = name
}

// WithKeyringSigner overrides the default keyring signer constructed
// by the node.
func WithKeyringSigner(signer *types.KeyringSigner) fx.Option {
	return fx.Replace(signer)
}
