package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/celestiaorg/celestia-node/nodebuilder"
)

var (
	coreFlag     = "core.ip"
	coreRPCFlag  = "core.rpc.port"
	coreGRPCFlag = "core.grpc.port"
)

// CoreFlags gives a set of hardcoded Core flags.
func CoreFlags() *flag.FlagSet {
	flags := &flag.FlagSet{}

	flags.String(
		coreFlag,
		"",
		"Indicates node to connect to the given core node. "+
			"Example: <ip>, 127.0.0.1. Assumes RPC port 26657 and gRPC port 9009 as default unless otherwise specified.",
	)
	flags.String(
		coreRPCFlag,
		"26657",
		"Set a custom RPC port for the core node connection. The --core.ip flag must also be provided.",
	)
	flags.String(
		coreGRPCFlag,
		"9090",
		"Set a custom gRPC port for the core node connection. The --core.ip flag must also be provided.",
	)
	return flags
}

// ParseCoreFlags parses Core flags from the given cmd and applies values to Env.
func ParseCoreFlags(
	ctx context.Context,
	cmd *cobra.Command,
	cfg *nodebuilder.Config,
) (setCtx context.Context, err error) {
	defer func() {
		setCtx = WithNodeConfig(ctx, cfg)
	}()
	coreIP := cmd.Flag(coreFlag).Value.String()
	if coreIP == "" {
		if cmd.Flag(coreGRPCFlag).Changed || cmd.Flag(coreRPCFlag).Changed {
			err = fmt.Errorf("cannot specify RPC/gRPC ports without specifying an IP address for --core.ip")
		}
		return
	}
	// sanity check given core ip addr and strip leading protocol
	ip, err := sanityCheckIP(coreIP)
	if err != nil {
		return
	}

	rpc := cmd.Flag(coreRPCFlag).Value.String()
	// sanity check rpc endpoint
	_, err = strconv.Atoi(rpc)
	if err != nil {
		return ctx, err
	}
	cfg.Core.SetRemoteCoreIP(ip)
	cfg.Core.SetRemoteCorePort(rpc)

	grpc := cmd.Flag(coreGRPCFlag).Value.String()
	// sanity check gRPC endpoint
	_, err = strconv.Atoi(grpc)
	if err != nil {
		return
	}
	cfg.Core.SetGRPCPort(grpc)
	return
}

// sanityCheckIP trims leading protocol scheme and port from the given
// IP address if present.
func sanityCheckIP(ip string) (string, error) {
	original := ip
	ip = strings.TrimPrefix(ip, "http://")
	ip = strings.TrimPrefix(ip, "https://")
	ip = strings.TrimPrefix(ip, "tcp://")
	ip = strings.TrimSuffix(ip, "/")
	ip = strings.Split(ip, ":")[0]
	if ip == "" {
		return "", fmt.Errorf("invalid IP addr given: %s", original)
	}
	return ip, nil
}
