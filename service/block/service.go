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

func NewBlockService(rpc *rpc.Client) (*Service, error) {
	if rpc == nil {
		return nil, fmt.Errorf("no running rpc found")
	}
	serv := &Service{
		rpc: rpc,
	}
	return serv, nil
}

func (s *Service) Start() error {

	return nil
}

func (s *Service) newBlockListener(ctx context.Context) error {
	newBlockEventChan, err := s.rpc.StartBlockSubscription(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <- ctx.Done():

		}
	}
}

// TODO:
// 	a function that has a listener loop for new events coming through the block subscription chan and then pipes them
//	 into the erasure coding function which then produces DAH, verifies against data root in header of raw block, if good
// 	gets stored and is ready to be served.
