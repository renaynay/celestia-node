package main

import (
	"github.com/spf13/cobra"

	"github.com/celestiaorg/celestia-node/cmd"
	"github.com/celestiaorg/celestia-node/node/node"
)

// NOTE: We should always ensure that the added Flags below are parsed somewhere, like in the PersistentPreRun func on
// parent command.

func init() {
	bridgeCmd.AddCommand(
		cmd.Init(
			cmd.NodeFlags(node.Bridge),
			cmd.P2PFlags(),
			cmd.CoreFlags(),
			cmd.MiscFlags(),
			cmd.RPCFlags(),
			cmd.KeyFlags(),
		),
		cmd.Start(
			cmd.NodeFlags(node.Bridge),
			cmd.P2PFlags(),
			cmd.CoreFlags(),
			cmd.MiscFlags(),
			cmd.RPCFlags(),
			cmd.KeyFlags(),
		),
	)
}

var bridgeCmd = &cobra.Command{
	Use:   "bridge [subcommand]",
	Args:  cobra.NoArgs,
	Short: "Manage your Bridge node",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var (
			ctx = cmd.Context()
			err error
		)

		ctx = cmd.WithNodeType(ctx, node.Bridge)

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
