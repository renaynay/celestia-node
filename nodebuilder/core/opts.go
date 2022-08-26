package core

import (
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/libs/fxutil"

	"github.com/celestiaorg/celestia-node/core"
)

type Option func(*settings)

// settings store values that can be augmented or changed for Node with Options.
type settings struct {
	cfg  *Config
	opts []fx.Option
}

// WithRemoteCoreIP configures Node to connect to the given remote Core IP.
func WithRemoteCoreIP(ip string) Option {
	return func(sets *settings) {
		sets.cfg.IP = ip
	}
}

// WithRemoteCorePort configures Node to connect to the given remote Core port.
func WithRemoteCorePort(port string) Option {
	return func(sets *settings) {
		sets.cfg.RPCPort = port
	}
}

// WithGRPCPort configures Node to connect to given gRPC port
// for state-related queries.
func WithGRPCPort(port string) Option {
	return func(sets *settings) {
		sets.cfg.GRPCPort = port
	}
}

// WithClient sets custom client for core process
func WithClient(client core.Client) Option {
	return func(sets *settings) {
		sets.opts = append(sets.opts, fxutil.ReplaceAs(client, new(core.Client)))
	}
}

// WithHeaderConstructFn sets custom func that creates extended header
func WithHeaderConstructFn(construct header.ConstructFn) Option {
	return func(sets *settings) {
		sets.opts = append(sets.opts, fx.Replace(construct))
	}
}
