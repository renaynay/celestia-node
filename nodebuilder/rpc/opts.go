package rpc

import (
	"github.com/celestiaorg/celestia-node/service/rpc"
)

// SetRPCPort configures Node to expose the given port for RPC
// queries.
// TODO(distractedm1nd): Once the config migrates to this package, define as a method
func SetRPCPort(cfg *rpc.Config, port string) {
	cfg.Port = port
}

// SetRPCAddress configures Node to listen on the given address for RPC
// queries.
// TODO(distractedm1nd): Once the config migrates to this package, define as a method
func SetRPCAddress(cfg *rpc.Config, addr string) {
	cfg.Address = addr
}
