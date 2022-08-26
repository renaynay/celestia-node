package cmd

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	nodebuilder "github.com/celestiaorg/celestia-node/node"

	"context"

	"github.com/celestiaorg/celestia-node/node/state"
)

var keyringAccNameFlag = "keyring.accname"

func KeyFlags() *flag.FlagSet {
	flags := &flag.FlagSet{}

	flags.String(keyringAccNameFlag, "", "Directs node's keyring signer to use the key prefixed with the "+
		"given string.")
	return flags
}

func ParseKeyFlags(ctx context.Context, cmd *cobra.Command) context.Context {
	keyringAccName := cmd.Flag(keyringAccNameFlag).Value.String()
	if keyringAccName != "" {
		return WithNodeOptions(ctx, nodebuilder.WithStateOptions(state.WithKeyringAccName(keyringAccName)))
	}
	return ctx
}
