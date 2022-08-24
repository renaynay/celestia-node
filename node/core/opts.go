package core

import "go.uber.org/fx"

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
