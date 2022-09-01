package core

import (
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/core"
	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/libs/fxutil"
)

// SetRemoteCoreIP configures Node to connect to the given remote Core IP.
func (cfg *Config) SetRemoteCoreIP(ip string) {
	cfg.IP = ip
}

// SetRemoteCorePort configures Node to connect to the given remote Core port.
func (cfg *Config) SetRemoteCorePort(port string) {
	cfg.RPCPort = port
}

// SetGRPCPort configures Node to connect to given gRPC port
// for state-related queries.
func (cfg *Config) SetGRPCPort(port string) {
	cfg.GRPCPort = port
}

// WithClient sets custom client for core process
func WithClient(client core.Client) fx.Option {
	return fxutil.ReplaceAs(client, new(core.Client))
}

// WithHeaderConstructFn sets custom func that creates extended header
func WithHeaderConstructFn(construct header.ConstructFn) fx.Option {
	return fx.Replace(construct)
}
