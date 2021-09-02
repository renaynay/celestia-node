package block

import (
	"context"
	"fmt"
	"github.com/celestiaorg/celestia-node/rpc"
)

// Service represents the Block service that can be started / stopped on a `Full` node.
// Service contains 4 main functionalities:
//		1. Fetching "raw" blocks from Celestia-Core. // TODO note: this will eventually take place on the p2p level.
// 		2. Erasure coding the "raw" blocks and producing a DataAvailabilityHeader + verifying the Data root.
// 		3. Storing erasure coded blocks.
// 		4. Serving erasure coded blocks to other `Full` node peers. // TODO note: optional for devnet
type Service struct {
	rpc *rpc.Client
}

// NewBlockService // TODO document
func NewBlockService(rpc *rpc.Client) (*Service, error) {
	if rpc == nil {
		return nil, fmt.Errorf("no running rpc found")
	}
	serv := &Service{
		rpc: rpc,
	}
	return serv, nil
}

// Start starts the BlockService.
func (s *Service) Start(ctx context.Context) error {
	return s.newBlockEventListener(ctx)
}
