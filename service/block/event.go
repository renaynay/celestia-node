package block

import (
	"context"
)

// listenForNewBlocks kicks of a listener loop that will
// listen for new "raw" blocks from the Fetcher, and handle
// them.
func (s *Service) listenForNewBlocks(ctx context.Context) error {
	// subscribe to new block events via the block fetcher
	newBlockEventChan, err := s.fetcher.SubscribeNewBlockEvent(ctx)
	if err != nil {
		return nil
	}
	// listen for new blocks from channel
	go func() {
		for {
			select {
			case <-s.stopListen:
				return
			case newRawBlock := <- newBlockEventChan:
				handleRawBlock(newRawBlock)
				// TODO @renaynay: how to handle errors here ?
				continue
			}
		}
	}()

	return nil
}

func handleRawBlock(raw *Raw) error {
	// TODO @renaynay:
	// extend the raw block
	// generate DAH using extended block
	// verify generated DAH against raw block data root
		// if fraud, generate fraud proof
	// store extended block

	return nil
}