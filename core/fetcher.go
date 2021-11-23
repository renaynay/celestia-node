package core

import (
	"context"
	"fmt"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/types"
)

const newBlockSubscriber = "NewBlock/Events"

var newBlockEventQuery = types.QueryForEvent(types.EventNewBlock).String()

type HeaderFetcher struct {
	client Client

	newBlockCh chan *types.Block
	doneCh     chan struct{}
}

// NewHeaderFetcher returns a new `HeaderFetcher`.
func NewHeaderFetcher(client Client) *HeaderFetcher {
	return &HeaderFetcher{
		client: client,
	}
}

// GetHeader queries Core for a `Block` at the given height and returns
// its Header.
func (f *HeaderFetcher) GetHeader(ctx context.Context, height *int64) (*types.Header, error) {
	res, err := f.client.Block(ctx, height)
	if err != nil {
		return nil, err
	}

	if res != nil && res.Block == nil {
		return nil, fmt.Errorf("core/fetcher: block not found")
	}
	header := res.Block.Header

	return &header, nil
}

// GetHeaderByHash queries Core for a `Block` at the given hash and returns
// its Header.
func (f *HeaderFetcher) GetHeaderByHash(ctx context.Context, hash tmbytes.HexBytes) (*types.Header, error) {
	res, err := f.client.BlockByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	if res != nil && res.Block == nil {
		return nil, fmt.Errorf("core/fetcher: block not found")
	}
	header := res.Block.Header

	return &header, nil
}

// Commit queries Core for a `Commit` from the block at
// the given height.
func (f *HeaderFetcher) Commit(ctx context.Context, height *int64) (*types.Commit, error) {
	res, err := f.client.Commit(ctx, height)
	if err != nil {
		return nil, err
	}

	if res != nil && res.Commit == nil {
		return nil, fmt.Errorf("core/fetcher: commit not found")
	}

	return res.Commit, nil
}

// ValidatorSet queries Core for the ValidatorSet from the
// block at the given height.
func (f *HeaderFetcher) ValidatorSet(ctx context.Context, height *int64) (*types.ValidatorSet, error) {
	res, err := f.client.Validators(ctx, height, nil, nil)
	if err != nil {
		return nil, err
	}

	if res != nil && res.Validators == nil {
		return nil, fmt.Errorf("core/fetcher: validators not found")
	}

	return types.NewValidatorSet(res.Validators), nil
}

// SubscribeNewHeaderEvent subscribes to new block events from Core, returning
// a new block event channel on success.
func (f *HeaderFetcher) SubscribeNewHeaderEvent(ctx context.Context) (<-chan *types.Block, error) {
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

	f.newBlockCh = make(chan *types.Block)
	f.doneCh = make(chan struct{})

	go func() {
		for {
			select {
			case <-f.doneCh:
				return
			case newEvent, ok := <-eventChan:
				if !ok {
					return
				}
				newBlock, ok := newEvent.Data.(types.EventDataNewBlock)
				if !ok {
					log.Warnf("unexpected event: %v", newEvent)
					continue
				}
				select {
				case f.newBlockCh <- newBlock.Block:
				case <-f.doneCh:
					return
				}
			}
		}
	}()

	return f.newBlockCh, nil
}

// UnsubscribeNewBlockEvent stops the subscription to new block events from Core.
func (f *HeaderFetcher) UnsubscribeNewBlockEvent(ctx context.Context) error {
	if f.newBlockCh == nil {
		return fmt.Errorf("no new block event channel found")
	}
	if f.doneCh == nil {
		return fmt.Errorf("no stop signal chan found in fetcher")
	}
	defer func() {
		// send stop signal
		f.doneCh <- struct{}{}
		// close out fetcher channels
		close(f.newBlockCh)
		close(f.doneCh)
		f.newBlockCh = nil
		f.doneCh = nil
	}()

	return f.client.Unsubscribe(ctx, newBlockSubscriber, newBlockEventQuery)
}
