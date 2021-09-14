package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedClientLifecycle(t *testing.T) {
	client := MockEmbeddedClient()
	require.NoError(t, client.Stop())
}

func TestEmbeddedClient_Status(t *testing.T) {
	client := MockEmbeddedClient()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	status, err := client.Status(ctx)
	require.NoError(t, err)
	assert.NotNil(t, status)

	require.NoError(t, client.Stop())
}

func TestEmbeddedClient_StartBlockSubscription_And_GetBlock(t *testing.T) {
	client := MockEmbeddedClient()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	eventChan, err := client.Subscribe(ctx, newBlockSubscriber, newBlockEventQuery)
	require.NoError(t, err)

	for i := 1; i <= 3; i++ {
		<-eventChan
		// check that `Block` works as intended (passing nil to get block at latest height)
		block, err := client.Block(ctx, nil)
		require.NoError(t, err)
		require.Equal(t, int64(i), block.Block.Height)
	}

	// unsubscribe to event channel
	require.NoError(t, client.Unsubscribe(ctx, newBlockSubscriber, newBlockEventQuery))
	require.NoError(t, client.Stop())
}

func TestRemoteClientLifecycle(t *testing.T) {
	remote, client, err := StartRemoteClient()
	require.NoError(t, err)

	require.NoError(t, client.Start())
	require.NoError(t, client.Stop())
	require.NoError(t, remote.Stop())
}

func TestRemoteClient_Status(t *testing.T) {
	remote, client, err := StartRemoteClient()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// client must be `started` before performing requests
	err = client.Start()
	require.NoError(t, err)

	status, err := client.Status(ctx)
	require.NoError(t, err)
	assert.NotNil(t, status)

	require.NoError(t, client.Stop())
	require.NoError(t, remote.Stop())
}

func TestRemoteClient_StartBlockSubscription_And_GetBlock(t *testing.T) {
	remote, client, err := StartRemoteClient()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// client must be `started` before subscribing to new block events.
	err = client.Start()
	require.NoError(t, err)

	eventChan, err := client.Subscribe(ctx, newBlockSubscriber, newBlockEventQuery)
	require.NoError(t, err)

	for i := 1; i <= 3; i++ {
		<-eventChan
		// check that `Block` works as intended (passing nil to get block at latest height)
		block, err := client.Block(ctx, nil)
		require.NoError(t, err)
		require.Equal(t, int64(i), block.Block.Height)
	}

	// unsubscribe to event channel
	require.NoError(t, client.Unsubscribe(ctx, newBlockSubscriber, newBlockEventQuery))
	require.NoError(t, client.Stop())
	require.NoError(t, remote.Stop())
}
