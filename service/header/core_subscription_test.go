package header

import (
	mdutils "github.com/ipfs/go-merkledag/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCoreSubscription(t *testing.T) {
	fetcher := createMockCoreFetcher()
	store := mdutils.Mock()

	// generate 10 blocks
	generateBlocks(t, fetcher)

	sub, err := newCoreSubscription(NewCoreExchange(fetcher, store))
	require.NoError(t, err)

	sub.NextHeader()
}

