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

	moduleOpts moduleOpts
}

type moduleOpts struct {
	state  []statemodule.Option
	header []headermodule.Option
	share  []sharemodule.Option
	rpc    []rpcmodule.Option
	p2p    []p2pmodule.Option
	core   []coremodule.Option
}

// WithStateOption is a top level option which allows customization for state module.
// NOTE: We have to make an option for each module for now as it is simple, though
// there are other ways of making this work via another level of indirection.
func WithStateOption(option statemodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.state = append(s.moduleOpts.state, option)
	}
}

// WithHeaderOption is a top level option which allows customization for header module.
// NOTE: See WithStateOption
func WithHeaderOption(option headermodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.header = append(s.moduleOpts.header, option)
	}
}

// WithShareOption is a top level option which allows customization for share module.
// NOTE: See WithStateOption
func WithShareOption(option sharemodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.share = append(s.moduleOpts.share, option)
	}
}

// WithRPCOption is a top level option which allows customization for rpc module.
// NOTE: See WithStateOption
func WithRPCOption(option rpcmodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.rpc = append(s.moduleOpts.rpc, option)
	}
}

// WithCoreOption is a top level option which allows customization for core module.
// NOTE: See WithStateOption
func WithCoreOption(option coremodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.core = append(s.moduleOpts.core, option)
	}
}

// WithP2POption is a top level option which allows customization for P2P module.
// NOTE: See WithStateOption
func WithP2POption(option p2pmodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.p2p = append(s.moduleOpts.p2p, option)
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
		sets.moduleOpts.header = append(sets.moduleOpts.header, headermodule.WithMetrics(enable))
	}
}
