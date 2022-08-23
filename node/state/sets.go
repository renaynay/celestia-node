package state

import (
	"go.uber.org/fx"
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
