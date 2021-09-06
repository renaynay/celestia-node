package block
// TODO consider renaming this file

import (
	"context"
	"github.com/celestiaorg/celestia-core/types"
)

// newBlockEventListener  // TODO document
func (s *Service) newBlockEventListener(ctx context.Context) <- chan error {
	// make channel to listen for block handling errors
	errCh := make(chan error)
	// subscribe to new block events via the block fetcher
	newBlockEventChan, err := s.fetcher.SubscribeNewBlockEvent(ctx)
	if err != nil {
		errCh <- err
		return errCh
	}
	// read once from new block event channel to ensure
	// the subscription has properly started // TODO why is this necessary?
	block := <- newBlockEventChan
	s.handleNewBlockEvent(block)
	// listen for new blocks
	go s.listen(ctx, newBlockEventChan, errCh)
	return errCh
}

// listen // TODO document
func (s *Service) listen(ctx context.Context, newBlockEventChan <-chan *types.Block, errCh chan error) {
	for {
		select {
		case <-ctx.Done():
			return
		case newBlock := <- newBlockEventChan:
			if err := s.handleNewBlockEvent(newBlock); err != nil {
				errCh <- err
			} // TODO should this be blocking?
		}
	}
}

// TODO eventually there will need to be some errChan
func (s *Service) handleNewBlockEvent(raw *types.Block) error {
	_, err := s.ErasureCodeBlock(raw)
	if err != nil {
		return err
	}

	// Generate DAH
	// TODO how does this interact with the HeaderService?

	// Verify DAH against Data Root? ---- `raw.DataAvailabilityHeader.Hash()`

	// Generate Fraud Proof (if bad encoding)
	// send fraud proof to peers

	// Store erasured block
	return nil
}

// TODO:
// 	a function that has a listener loop for new events coming through the block subscription chan and then pipes them
//	 into the erasure coding function which then produces DAH, verifies against data root in header of raw block, if good
// 	gets stored and is ready to be served.
