package cmd

import (
	"context"

	rpcmodule "github.com/celestiaorg/celestia-node/nodebuilder/rpc"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/celestiaorg/celestia-node/nodebuilder"
)

var (
	addrFlag = "rpc.addr"
	portFlag = "rpc.port"
)

// RPCFlags gives a set of hardcoded node/rpc package flags.
func RPCFlags() *flag.FlagSet {
	flags := &flag.FlagSet{}

	flags.String(
		addrFlag,
		"",
		"Set a custom RPC listen address (default: localhost)",
	)
	flags.String(
		portFlag,
		"",
		"Set a custom RPC port (default: 26658)",
	)

	return flags
}

// ParseRPCFlags parses RPC flags from the given cmd and applies values to Env.
func ParseRPCFlags(ctx context.Context, cmd *cobra.Command, cfg *nodebuilder.Config) (context.Context, error) {
	addr := cmd.Flag(addrFlag).Value.String()
	if addr != "" {
		rpcmodule.SetRPCAddress(&cfg.RPC, addr)
		ctx = WithNodeConfig(ctx, cfg)
	}
	port := cmd.Flag(portFlag).Value.String()
	if port != "" {
		rpcmodule.SetRPCPort(&cfg.RPC, port)
		ctx = WithNodeConfig(ctx, cfg)
	}
	return ctx, nil
}
