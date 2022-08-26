package rpc

import (
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/service/rpc"
)

type Option func(*settings)

// settings store values that can be augmented or changed for Node with Options.
type settings struct {
	cfg  *rpc.Config
	opts []fx.Option
}

// WithRPCPort configures Node to expose the given port for RPC
// queries.
func WithRPCPort(port string) Option {
	return func(sets *settings) {
		sets.cfg.Port = port
	}
}

// WithRPCAddress configures Node to listen on the given address for RPC
// queries.
func WithRPCAddress(addr string) Option {
	return func(sets *settings) {
		sets.cfg.Address = addr
	}
}

func WithOption(option fx.Option) Option {
	return func(sets *settings) {
		sets.opts = append(sets.opts, option)
	}
}
