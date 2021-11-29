package header

import (
	"context"
	mdutils "github.com/ipfs/go-merkledag/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCoreSubscription(t *testing.T) {
	fetcher := createMockCoreFetcher()
	store := mdutils.Mock()

	sub, err := newCoreSubscription(NewCoreExchange(fetcher, store))
	require.NoError(t, err)

	// generate 10 blocks
	generateBlocks(sub.sub, 10)
	// ensure they can be read from the channel
	for i := 0; i < 10; i++ {
		next, err := sub.NextHeader(context.Background())
		require.NoError(t, err)
		require.NotNil(t, next)
	}
}

