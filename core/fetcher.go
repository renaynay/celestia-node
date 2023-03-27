package core

import (
	"context"
	"fmt"
	"strings"

	logging "github.com/ipfs/go-log/v2"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

const newBlockSubscriber = "NewBlock/Events"

var (
	log                     = logging.Logger("core")
	newDataSignedBlockQuery = types.QueryForEvent(types.EventSignedBlock).String()
)

type BlockFetcher struct {
	client Client

	chainID string

	doneCh chan struct{}
	cancel context.CancelFunc
}

// NewBlockFetcher returns a new `BlockFetcher`.
func NewBlockFetcher(client Client, chainID string) *BlockFetcher {
	return &BlockFetcher{
		client:  client,
		chainID: chainID,
	}
}

// GetSignedBlock queries Core for a `Block` at the given height.
func (f *BlockFetcher) GetSignedBlock(ctx context.Context, height *int64) (*coretypes.ResultSignedBlock, error) {
	signed, err := f.client.SignedBlock(ctx, height)
	if err != nil {
		return nil, err
	}
	return signed, f.validateChainID(signed.Header.ChainID)
}

// Commit queries Core for a `Commit` from the block at
// the given height.
func (f *BlockFetcher) Commit(ctx context.Context, height *int64) (*types.Commit, error) {
	res, err := f.client.Commit(ctx, height)
	if err != nil {
		return nil, err
	}

	if res != nil && res.Commit == nil {
		return nil, fmt.Errorf("core/fetcher: commit not found at height %d", height)
	}

	return res.Commit, nil
}

// ValidatorSet queries Core for the ValidatorSet from the
// block at the given height.
func (f *BlockFetcher) ValidatorSet(ctx context.Context, height *int64) (*types.ValidatorSet, error) {
	var perPage = 100

	vals, total := make([]*types.Validator, 0), -1
	for page := 1; len(vals) != total; page++ {
		res, err := f.client.Validators(ctx, height, &page, &perPage)
		if err != nil {
			return nil, err
		}

		if res != nil && len(res.Validators) == 0 {
			return nil, fmt.Errorf("core/fetcher: validator set not found at height %d", height)
		}

		total = res.Total
		vals = append(vals, res.Validators...)
	}

	return types.NewValidatorSet(vals), nil
}

// SubscribeNewBlockEvent subscribes to new block events from Core, returning
// a new block event channel on success.
func (f *BlockFetcher) SubscribeNewBlockEvent(ctx context.Context) (<-chan types.EventDataSignedBlock, error) {
	// start the client if not started yet
	if !f.client.IsRunning() {
		return nil, fmt.Errorf("client not running")
	}

	ctx, cancel := context.WithCancel(ctx)
	f.cancel = cancel
	f.doneCh = make(chan struct{})

	eventChan, err := f.client.Subscribe(ctx, newBlockSubscriber, newDataSignedBlockQuery)
	if err != nil {
		return nil, err
	}

	signedBlockCh := make(chan types.EventDataSignedBlock)
	go func() {
		defer close(f.doneCh)
		defer close(signedBlockCh)
		for {
			select {
			case <-ctx.Done():
				return
			case newEvent, ok := <-eventChan:
				if !ok {
					log.Errorw("fetcher: new blocks subscription channel closed unexpectedly")
					return
				}

				signedBlock := newEvent.Data.(types.EventDataSignedBlock)
				if err := f.validateChainID(signedBlock.Header.ChainID); err != nil {
					log.Errorw("fetcher: received block with unexpected chainID: expected %s, got %s",
						f.chainID,
						signedBlock.Header.ChainID,
					)
					return
				}

				select {
				case signedBlockCh <- signedBlock:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return signedBlockCh, nil
}

// UnsubscribeNewBlockEvent stops the subscription to new block events from Core.
func (f *BlockFetcher) UnsubscribeNewBlockEvent(ctx context.Context) error {
	f.cancel()
	select {
	case <-f.doneCh:
	case <-ctx.Done():
		return fmt.Errorf("fetcher: unsubscribe from new block events: %w", ctx.Err())
	}
	return f.client.Unsubscribe(ctx, newBlockSubscriber, newDataSignedBlockQuery)
}

// IsSyncing returns the sync status of the Core connection: true for
// syncing, and false for already caught up. It can also return an error
// in the case of a failed status request.
func (f *BlockFetcher) IsSyncing(ctx context.Context) (bool, error) {
	resp, err := f.client.Status(ctx)
	if err != nil {
		return false, err
	}
	return resp.SyncInfo.CatchingUp, nil
}

// validateChainID returns an error if there is a chainID mismatch.
func (f *BlockFetcher) validateChainID(chainID string) error {
	if !strings.EqualFold(f.chainID, chainID) {
		return fmt.Errorf("header with different chainID received: expected %s, got %s",
			f.chainID, chainID,
		)
	}
	return nil
}
