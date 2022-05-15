//nolint:dupl
package main

import (
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/spf13/cobra"

	cmdnode "github.com/celestiaorg/celestia-node/cmd"
	"github.com/celestiaorg/celestia-node/node"
)

// NOTE: We should always ensure that the added Flags below are parsed somewhere, like in the PersistentPreRun func on
// parent command.

func init() {
	lightKeyCmd := keys.Commands("~/.celestia-light/keys")
	lightKeyCmd.Short = "Manage your light node account keys"
	lightKeyCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		_, err := cmdnode.GetEnv(cmd.Context())
		return err
	}

	lightCmd.AddCommand(
		cmdnode.Init(
			cmdnode.NodeFlags(node.Light),
			cmdnode.P2PFlags(),
			cmdnode.HeadersFlags(),
			cmdnode.MiscFlags(),
			// NOTE: for now, state-related queries can only be accessed
			// over an RPC connection with a celestia-core node.
			cmdnode.CoreFlags(),
			cmdnode.RPCFlags(),
			cmdnode.KeyFlags(),
		),
		cmdnode.Start(
			cmdnode.NodeFlags(node.Light),
			cmdnode.P2PFlags(),
			cmdnode.HeadersFlags(),
			cmdnode.MiscFlags(),
			// NOTE: for now, state-related queries can only be accessed
			// over an RPC connection with a celestia-core node.
			cmdnode.CoreFlags(),
			cmdnode.RPCFlags(),
			cmdnode.KeyFlags(),
		),
		lightKeyCmd,
	)
}

var lightCmd = &cobra.Command{
	Use:   "light [subcommand]",
	Args:  cobra.NoArgs,
	Short: "Manage your Light node",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		env, err := cmdnode.GetEnv(cmd.Context())
		if err != nil {
			return err
		}
		env.SetNodeType(node.Light)

		err = cmdnode.ParseNodeFlags(cmd, env)
		if err != nil {
			return err
		}

		err = cmdnode.ParseP2PFlags(cmd, env)
		if err != nil {
			return err
		}

		err = cmdnode.ParseCoreFlags(cmd, env)
		if err != nil {
			return err
		}

		err = cmdnode.ParseHeadersFlags(cmd, env)
		if err != nil {
			return err
		}

		err = cmdnode.ParseMiscFlags(cmd)
		if err != nil {
			return err
		}

		err = cmdnode.ParseRPCFlags(cmd, env)
		if err != nil {
			return err
		}

		cmdnode.ParseKeyFlags(cmd, env)

		return nil
	},
}
