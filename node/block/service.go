package block

import (
	"github.com/celestiaorg/celestia-node/node/rpc"
	"github.com/celestiaorg/celestia-node/service/block"
)

type Config struct {
	// TODO
}

func Components(rpc *rpc.Client) interface{} {
	return func() (*block.Service, error) {
		return block.NewBlockService(rpc)
	}
}
