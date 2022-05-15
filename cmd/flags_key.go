package cmd

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/celestiaorg/celestia-node/node"
)

var (
	keyNameFlag    = "key.name"
	keyBackendFlag = "key.backend"
)

// KeyFlags gives a set of key-related flags.
func KeyFlags() *flag.FlagSet {
	flags := &flag.FlagSet{}

	flags.String(keyNameFlag, "",
		"Indicates key name to use as the node's default account. (default: \"celes\")")
	flags.String(keyBackendFlag, "",
		"Indicates keyring-backend to use for keyring construction. (default: \"file\")")
	return flags
}

// ParseKeyFlags parses key-related flags from the given cmd and applies values to Env.
func ParseKeyFlags(cmd *cobra.Command, env *Env) {
	keyName := cmd.Flag(keyNameFlag).Value.String()
	if keyName != "" {
		env.AddOptions(node.WithKeyName(keyName))
	}

	keyBackend := cmd.Flag(keyBackendFlag).Value.String()
	if keyBackend != "" {
		env.AddOptions(node.WithKeyBackend(keyBackend))
	}
}
