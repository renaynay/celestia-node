package cmd

import (
	"github.com/celestiaorg/celestia-node/node"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var (
	keyNameFlag    = "key.name"
	keyBackendFlag = "key.backend"
)

// KeyFlags gives a set of key-related flags.
func KeyFlags() *flag.FlagSet {
	flags := &flag.FlagSet{}

	flags.String(keyNameFlag, "celes",
		"Indicates key name to use as the node's default account. (default: \"celes\")")
	flags.String(keyBackendFlag, "os",
		"Indicates keyring-backend to use for keyring construction. (default: \"os\")")
	return flags
}

// ParseKeyFlags parses key-related flags from the given cmd and applies values to Env.
func ParseKeyFlags(cmd *cobra.Command, env *Env) error {
	keyName := cmd.Flag(keyNameFlag).Value.String()
	env.AddOptions(node.WithKeyName(keyName))

	keyBackend := cmd.Flag(keyBackendFlag).Value.String()
	env.AddOptions(node.WithKeyBackend(keyBackend))
	return nil
}
