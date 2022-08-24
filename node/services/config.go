package services

import (
	"time"

	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("node/services")

type Config struct {
	// NOTE: All further fields related to share/discovery.
	// PeersLimit defines how many peers will be added during discovery.
	PeersLimit uint
	// DiscoveryInterval is an interval between discovery sessions.
	DiscoveryInterval time.Duration
	// AdvertiseInterval is a interval between advertising sessions.
	// NOTE: only full and bridge can advertise themselves.
	AdvertiseInterval time.Duration
}

func DefaultConfig() Config {
	return Config{
		PeersLimit:        3,
		DiscoveryInterval: time.Second * 30,
		AdvertiseInterval: time.Second * 30,
	}
}
