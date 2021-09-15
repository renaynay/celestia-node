package core

import (
	"context"
	"fmt"
	"sync"

	rpctypes "github.com/celestiaorg/celestia-core/rpc/core/types"
	"github.com/celestiaorg/celestia-core/types"
	"github.com/celestiaorg/celestia-node/service/block"
)

const newBlockSubscriber = "NewBlock/Events"

var newBlockEventQuery = types.QueryForEvent(types.EventNewBlock).String()

type BlockFetcher struct {
	client Client

	mux        sync.Mutex
	newBlockCh chan *block.Raw
}

// NewBlockFetcher returns a new `BlockFetcher`.
func NewBlockFetcher(client Client) *BlockFetcher {
	return &BlockFetcher{
		client: client,
	}
}

// GetBlock queries Core for a `Block` at the given height.
func (f *BlockFetcher) GetBlock(ctx context.Context, height *int64) (*block.Raw, error) {
	raw, err := f.client.Block(ctx, height)
	if err != nil {
		return nil, err
	}
	return raw.Block, nil
}

// SubscribeNewBlockEvent subscribes to new block events from Core, returning
// a new block event channel on success.
func (f *BlockFetcher) SubscribeNewBlockEvent(ctx context.Context) (<-chan *block.Raw, error) {
	// start the client if not started yet
	if !f.client.IsRunning() {
		return nil, fmt.Errorf("client not running")
	}
	eventChan, err := f.client.Subscribe(ctx, newBlockSubscriber, newBlockEventQuery)
	if err != nil {
		return nil, err
	}

	// create a wrapper channel for translating ResultEvent to "raw" block
	if f.newBlockCh != nil {
		return nil, fmt.Errorf("new block event channel exists")
	}

	newBlockChan := make(chan *block.Raw)
	f.mux.Lock()
	f.newBlockCh = newBlockChan
	f.mux.Unlock()

	go func(eventChan <-chan rpctypes.ResultEvent, newBlockChan chan *block.Raw) {
		for {
			newEvent := <-eventChan
			newBlock, ok := newEvent.Data.(types.EventDataNewBlock)
			if !ok {
				// TODO @renaynay: log & ignore
				continue
			}
			newBlockChan <- newBlock.Block
		}
	}(eventChan, newBlockChan)

	return newBlockChan, nil
}

// UnsubscribeNewBlockEvent stops the subscription to new block events from Core.
func (f *BlockFetcher) UnsubscribeNewBlockEvent(ctx context.Context) error {
	// close the new block channel
	if f.newBlockCh == nil {
		return fmt.Errorf("no new block event channel found")
	}
	close(f.newBlockCh)
	f.mux.Lock()
	f.newBlockCh = nil
	f.mux.Unlock()

	return f.client.Unsubscribe(ctx, newBlockSubscriber, newBlockEventQuery)
}
