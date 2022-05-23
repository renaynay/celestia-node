package cmd

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/celestiaorg/celestia-node/node"
)

var (
	enableTraceFlag = "metrics.trace"
)

func MetricsFlags() *flag.FlagSet {
	flags := &flag.FlagSet{}
	flags.Bool(enableTraceFlag, false, "Enables tracing for services provided by celestia-node")

	return flags
}

func ParseMetricsFlags(cmd *cobra.Command, env *Env) error {
	enableTrace, err := cmd.Flags().GetBool(enableTraceFlag)
	if err != nil {
		return err
	}
	if enableTrace {
		env.AddOptions(node.WithTracingEnabled())
	}
	return nil
}
