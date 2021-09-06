package block

import (
	"context"
)

// Service represents the Block service that can be started / stopped on a `Full` node.
// Service contains 4 main functionalities:
//		1. Fetching "raw" blocks from Celestia-Core. // TODO note: this will eventually take place on the p2p level.
// 		2. Erasure coding the "raw" blocks and producing a DataAvailabilityHeader + verifying the Data root.
// 		3. Storing erasure coded blocks.
// 		4. Serving erasure coded blocks to other `Full` node peers. // TODO note: optional for devnet
type Service struct {
	fetcher BlockFetcher

	// TODO should I store all the channels on the Service ? would make it easier to organise.
}

// NewBlockService creates a new instance of block Service.
func NewBlockService(fetcher BlockFetcher) *Service {
	return &Service{
		fetcher: fetcher,
	}
}

// Start starts the block Service.
func (s *Service) Start(ctx context.Context) <- chan error { // TODO this chan needs to be closed somehow
	return s.newBlockEventListener(ctx)
}

// TODO
func (s *Service) Stop(ctx context.Context) error {
	// TODO make this more robust
	ctx.Done()
	return s.fetcher.UnsubscribeNewBlockEvent(ctx)
}

