package nodebuilder

import (
	"go.uber.org/fx"

	coremodule "github.com/celestiaorg/celestia-node/nodebuilder/core"
	headermodule "github.com/celestiaorg/celestia-node/nodebuilder/header"
	p2pmodule "github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	rpcmodule "github.com/celestiaorg/celestia-node/nodebuilder/rpc"
	sharemodule "github.com/celestiaorg/celestia-node/nodebuilder/share"
	statemodule "github.com/celestiaorg/celestia-node/nodebuilder/state"
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

// WithStateOptions is a top level option which allows customization for state module.
// NOTE: We have to make an option for each module for now as it is simple, though
// there are other ways of making this work via another level of indirection.
func WithStateOptions(options ...statemodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.state = append(s.moduleOpts.state, options...)
	}
}

// WithHeaderOptions is a top level option which allows customization for header module.
// NOTE: See WithStateOptions
func WithHeaderOptions(options ...headermodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.header = append(s.moduleOpts.header, options...)
	}
}

// WithShareOptions is a top level option which allows customization for share module.
// NOTE: See WithStateOptions
func WithShareOptions(options ...sharemodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.share = append(s.moduleOpts.share, options...)
	}
}

// WithRPCOptions is a top level option which allows customization for rpc module.
// NOTE: See WithStateOptions
func WithRPCOptions(options ...rpcmodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.rpc = append(s.moduleOpts.rpc, options...)
	}
}

// WithCoreOptions is a top level option which allows customization for core module.
// NOTE: See WithStateOptions
func WithCoreOptions(options ...coremodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.core = append(s.moduleOpts.core, options...)
	}
}

// WithP2pOptions is a top level option which allows customization for P2P module.
// NOTE: See WithStateOptions
func WithP2pOptions(options ...p2pmodule.Option) Option {
	return func(s *settings) {
		s.moduleOpts.p2p = append(s.moduleOpts.p2p, options...)
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
