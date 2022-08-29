package header

import (
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/header"
)

// SetTrustedHash sets TrustedHash to the Config.
func (cfg *Config) SetTrustedHash(hash string) {
	cfg.TrustedHash = hash
}

// AddTrustedPeers appends new "trusted peers" to the Config.
func (cfg *Config) AddTrustedPeers(addr ...string) {
	cfg.TrustedPeers = append(cfg.TrustedPeers, addr...)
}

// TODO: Eventually we should have a per-module metrics option.
// WithMetrics enables metrics exporting for the node.
func WithMetrics(enable bool) fx.Option {
	if !enable {
		return fx.Options()
	}
	return fx.Options(fx.Invoke(header.MonitorHead))
}
