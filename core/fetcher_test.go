package core

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/types"
)

func TestBlockFetcher_GetBlock_and_SubscribeNewBlockEvent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	t.Cleanup(cancel)

	client := StartTestNode(t).Client
	fetcher := NewBlockFetcher(client, "private")

	// generate some blocks
	newBlockChan, err := fetcher.SubscribeNewBlockEvent(ctx)
	require.NoError(t, err)

	for i := 1; i < 3; i++ {
		select {
		case newBlockFromChan := <-newBlockChan:
			h := newBlockFromChan.Header.Height
			block, err := fetcher.GetSignedBlock(ctx, &h)
			require.NoError(t, err)
			assert.Equal(t, newBlockFromChan.Data, block.Data)
			assert.Equal(t, newBlockFromChan.Header, block.Header)
			assert.Equal(t, newBlockFromChan.Commit, block.Commit)
			assert.Equal(t, newBlockFromChan.ValidatorSet, block.ValidatorSet)
			require.GreaterOrEqual(t, newBlockFromChan.Header.Height, int64(i))
		case <-ctx.Done():
			require.NoError(t, ctx.Err())
		}
	}
	require.NoError(t, fetcher.UnsubscribeNewBlockEvent(ctx))
}

// TestBlockFetcherHeaderValues tests that both the Commit and ValidatorSet
// endpoints are working as intended.
func TestBlockFetcherHeaderValues(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	t.Cleanup(cancel)

	client := StartTestNode(t).Client
	fetcher := NewBlockFetcher(client, "private")

	// generate some blocks
	newBlockChan, err := fetcher.SubscribeNewBlockEvent(ctx)
	require.NoError(t, err)
	// read once from channel to generate next block
	var h int64
	select {
	case evt := <-newBlockChan:
		h = evt.Header.Height
	case <-ctx.Done():
		require.NoError(t, ctx.Err())
	}
	firstBlock, err := fetcher.GetSignedBlock(ctx, &h)
	require.NoError(t, err)
	// get next block
	var nextBlock types.EventDataSignedBlock
	select {
	case nextBlock = <-newBlockChan:
	case <-ctx.Done():
		require.NoError(t, ctx.Err())
	}
	// compare LastCommit from next block to Commit from first block height
	assert.Equal(t, nextBlock.Header.LastCommitHash, firstBlock.Commit.Hash())
	assert.Equal(t, nextBlock.Header.Height, firstBlock.Commit.Height+1)
	// compare ValidatorSet hash to the ValidatorsHash from first block height
	assert.Equal(t, nextBlock.ValidatorSet.Hash(), firstBlock.ValidatorSet.Hash())
	require.NoError(t, fetcher.UnsubscribeNewBlockEvent(ctx))
}
