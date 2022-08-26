package header

import (
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/header"
)

type Option func(*settings)

// settings store values that can be augmented or changed for Node with Options.
type settings struct {
	cfg  *Config
	opts []fx.Option
}

// WithTrustedHash sets TrustedHash to the Config.
func WithTrustedHash(hash string) Option {
	return func(sets *settings) {
		sets.cfg.TrustedHash = hash
	}
}

// WithTrustedPeers appends new "trusted peers" to the Config.
func WithTrustedPeers(addr ...string) Option {
	return func(sets *settings) {
		sets.cfg.TrustedPeers = append(sets.cfg.TrustedPeers, addr...)
	}
}

// TODO: Eventually we should have a per-module metrics option.
// WithMetrics enables metrics exporting for the node.
func WithMetrics(enable bool) Option {
	return func(sets *settings) {
		if !enable {
			return
		}
		sets.opts = append(sets.opts, fx.Options(fx.Invoke(header.MonitorHead)))
	}
}
