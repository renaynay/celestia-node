package share

import (
	"time"
)

// SetPeersLimit overrides default peer limit for peers found during discovery.
func (cfg *Config) SetPeersLimit(limit uint) {
	cfg.PeersLimit = limit
}

// SetDiscoveryInterval sets interval between discovery sessions.
func (cfg *Config) SetDiscoveryInterval(interval time.Duration) {
	if interval <= 0 {
		return
	}
	cfg.DiscoveryInterval = interval
}

// SetAdvertiseInterval sets interval between advertises.
func (cfg *Config) SetAdvertiseInterval(interval time.Duration) {
	if interval <= 0 {
		return
	}
	cfg.AdvertiseInterval = interval
}
