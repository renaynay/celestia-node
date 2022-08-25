package node

import (
	"go.uber.org/fx"

	coremodule "github.com/celestiaorg/celestia-node/node/core"
	headermodule "github.com/celestiaorg/celestia-node/node/header"
	p2pmodule "github.com/celestiaorg/celestia-node/node/p2p"
	rpcmodule "github.com/celestiaorg/celestia-node/node/rpc"
	sharemodule "github.com/celestiaorg/celestia-node/node/share"
	statemodule "github.com/celestiaorg/celestia-node/node/state"
	"github.com/celestiaorg/celestia-node/params"
)

// settings store values that can be augmented or changed for Node with Options.
type settings struct {
	cfg  *Config
	opts []fx.Option

	moduleOpts ModuleOpts
}

type ModuleOpts struct {
	State  []statemodule.Option
	Header []headermodule.Option
	Share  []sharemodule.Option
	RPC    []rpcmodule.Option
	P2P    []p2pmodule.Option
	Core   []coremodule.Option
}

// WithStateOption is a top level option which allows customization for state module.
// NOTE: We have to make an option for each module for now as it is simple, though
// there are other ways of making this work via another level of indirection.
func WithStateOption(option statemodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.State = append(s.moduleOpts.State, option)
	}
}

// WithHeaderOption is a top level option which allows customization for header module.
// NOTE: See WithStateOption
func WithHeaderOption(option headermodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.Header = append(s.moduleOpts.Header, option)
	}
}

// WithShareOption is a top level option which allows customization for share module.
// NOTE: See WithStateOption
func WithShareOption(option sharemodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.Share = append(s.moduleOpts.Share, option)
	}
}

// WithRPCOption is a top level option which allows customization for rpc module.
// NOTE: See WithStateOption
func WithRPCOption(option rpcmodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.RPC = append(s.moduleOpts.RPC, option)
	}
}

// WithCoreOption is a top level option which allows customization for core module.
// NOTE: See WithStateOption
func WithCoreOption(option coremodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.Core = append(s.moduleOpts.Core, option)
	}
}

// WithP2POption is a top level option which allows customization for P2P module.
// NOTE: See WithStateOption
func WithP2POption(option p2pmodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.P2P = append(s.moduleOpts.P2P, option)
	}
}

// Option for Node's Config.
type Option func(*settings)

// WithNetwork specifies the Network to which the Node should connect to.
// WARNING: Use this option with caution and never run the Node with different networks over the same persisted Store.
func WithNetwork(net params.Network) Option {
	return func(sets *settings) {
		sets.opts = append(sets.opts, fx.Replace(net))
	}
}

// WithBootstrappers sets custom bootstrap peers.
func WithBootstrappers(peers params.Bootstrappers) Option {
	return func(sets *settings) {
		sets.opts = append(sets.opts, fx.Replace(peers))
	}
}

// WithMetrics enables metrics exporting for the node.
func WithMetrics(enable bool) Option {
	return func(sets *settings) {
		if !enable {
			return
		}
		sets.moduleOpts.Header = append(sets.moduleOpts.Header, headermodule.WithMetrics(enable))
	}
}
