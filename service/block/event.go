package block

import (
	"context"

	"github.com/celestiaorg/celestia-core/types"
)

// TODO consider renaming this file

func (s *Service) newBlockEventListener(ctx context.Context) error {
	newBlockEventChan, err := s.rpc.StartBlockSubscription(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case newBlockEvent := <- newBlockEventChan:
			rawBlock, ok := newBlockEvent.Data.(types.EventDataNewBlock)
			if !ok {
				// TODO log error or return err? Nothing non-block related should come thru this pipe
				continue
			}
			s.handleNewBlockEvent(rawBlock.Block) // TODO sync or async?
		case <- ctx.Done():
			return nil
		}
	}
}

func (s *Service) handleNewBlockEvent(raw *types.Block) {
	_, err := s.ErasureCodeBlock(raw)
	if err != nil {
		// TODO handle error
	}

	// Generate DAH
		// TODO how does this interact with the HeaderService?

	// Verify DAH against Data Root? ---- `raw.DataAvailabilityHeader.Hash()`

	// Generate Fraud Proof (if bad encoding)
		// send fraud proof to peers

	// Store erasured block
}

// TODO:
// 	a function that has a listener loop for new events coming through the block subscription chan and then pipes them
//	 into the erasure coding function which then produces DAH, verifies against data root in header of raw block, if good
// 	gets stored and is ready to be served.

