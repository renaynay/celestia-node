package core

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	ds "github.com/ipfs/go-datastore"
	ds_sync "github.com/ipfs/go-datastore/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmrand "github.com/tendermint/tendermint/libs/rand"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/share/eds"
)

func TestCoreExchange_RequestHeaders(t *testing.T) {
	fetcher := createCoreFetcher(t)

	// generate 10 blocks
	generateBlocks(t, fetcher)

	store := createStore(t)

	ce := NewExchange(fetcher, store, header.MakeExtendedHeader)
	headers, err := ce.GetRangeByHeight(context.Background(), 1, 10)
	require.NoError(t, err)

	assert.Equal(t, 10, len(headers))
}

func createCoreFetcher(t *testing.T) *BlockFetcher {
	cfg := DefaultTestConfig()
	cfg.Accounts = []string{tmrand.Str(9)}
	cctx := StartTestNodeWithConfig(t, cfg)
	for i := 0; i < 6; i++ {
		_, err := cctx.FillBlock(8, cfg.Accounts, flags.BroadcastBlock)
		require.NoError(t, err)
	}

	return NewBlockFetcher(cctx.Client)
}

func createStore(t *testing.T) *eds.Store {
	store, err := eds.NewStore(t.TempDir(), ds_sync.MutexWrap(ds.NewMapDatastore()))
	require.NoError(t, err)
	return store
}

func generateBlocks(t *testing.T, fetcher *BlockFetcher) {
	sub, err := fetcher.SubscribeNewBlockEvent(context.Background())
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		<-sub
	}
}
