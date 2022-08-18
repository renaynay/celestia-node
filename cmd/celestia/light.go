//nolint:dupl
package main

import (
	"github.com/spf13/cobra"

	"github.com/celestiaorg/celestia-node/cmd"
	"github.com/celestiaorg/celestia-node/node/node"
)

// NOTE: We should always ensure that the added Flags below are parsed somewhere, like in the PersistentPreRun func on
// parent command.

func init() {
	lightCmd.AddCommand(
		cmd.Init(
			cmd.NodeFlags(node.Light),
			cmd.P2PFlags(),
			cmd.HeadersFlags(),
			cmd.MiscFlags(),
			// NOTE: for now, state-related queries can only be accessed
			// over an RPC connection with a celestia-core node.
			cmd.CoreFlags(),
			cmd.RPCFlags(),
			cmd.KeyFlags(),
		),
		cmd.Start(
			cmd.NodeFlags(node.Light),
			cmd.P2PFlags(),
			cmd.HeadersFlags(),
			cmd.MiscFlags(),
			// NOTE: for now, state-related queries can only be accessed
			// over an RPC connection with a celestia-core node.
			cmd.CoreFlags(),
			cmd.RPCFlags(),
			cmd.KeyFlags(),
		),
	)
}

var lightCmd = &cobra.Command{
	Use:   "light [subcommand]",
	Args:  cobra.NoArgs,
	Short: "Manage your Light node",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var (
			ctx = cmd.Context()
			err error
		)

		ctx = cmd.WithNodeType(ctx, node.Light)

		ctx, err = cmd.ParseNodeFlags(ctx, cmd)
		if err != nil {
			return err
		}

		ctx, err = cmd.ParseP2PFlags(ctx, cmd)
		if err != nil {
			return err
		}

		ctx, err = cmd.ParseCoreFlags(ctx, cmd)
		if err != nil {
			return err
		}

		ctx, err = cmd.ParseHeadersFlags(ctx, cmd)
		if err != nil {
			return err
		}

		ctx, err = cmd.ParseMiscFlags(ctx, cmd)
		if err != nil {
			return err
		}

		ctx, err = cmd.ParseRPCFlags(ctx, cmd)
		if err != nil {
			return err
		}

		ctx = cmd.ParseKeyFlags(ctx, cmd)

		cmd.SetContext(ctx)
		return nil
	},
}
