package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	core2 "github.com/celestiaorg/celestia-node/node/core"

	"github.com/celestiaorg/celestia-node/core"
	"github.com/celestiaorg/celestia-node/node/node"
	"github.com/celestiaorg/celestia-node/params"
)

func TestBridge_WithMockedCoreClient(t *testing.T) {
	t.Skip("skipping") // consult https://github.com/celestiaorg/celestia-core/issues/667 for reasoning
	repo := MockStore(t, DefaultConfig(node.Bridge))

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	_, client := core.StartTestClient(ctx, t)
	node, err := New(node.Bridge, repo,
		WithCoreOption(core2.WithClient(client)),
		WithNetwork(params.Private),
	)
	require.NoError(t, err)
	require.NotNil(t, node)
	err = node.Start(ctx)
	require.NoError(t, err)

	err = node.Stop(ctx)
	require.NoError(t, err)
}
