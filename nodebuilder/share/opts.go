package share

import (
	"time"

	"go.uber.org/fx"
)

type Option func(*settings)

// settings store values that can be augmented or changed for Node with Options.
type settings struct {
	cfg  *Config
	opts []fx.Option
}

// WithPeersLimit overrides default peer limit for peers found during discovery.
func WithPeersLimit(limit uint) Option {
	return func(sets *settings) {
		sets.cfg.PeersLimit = limit
	}
}

// WithDiscoveryInterval sets interval between discovery sessions.
func WithDiscoveryInterval(interval time.Duration) Option {
	return func(sets *settings) {
		if interval <= 0 {
			return
		}
		sets.cfg.DiscoveryInterval = interval
	}
}

// WithAdvertiseInterval sets interval between advertises.
func WithAdvertiseInterval(interval time.Duration) Option {
	return func(sets *settings) {
		if interval <= 0 {
			return
		}
		sets.cfg.AdvertiseInterval = interval
	}
}
