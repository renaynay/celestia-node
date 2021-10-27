package header

import (
	"context"
	"testing"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscribe(t *testing.T) {
	net, err := mocknet.FullMeshConnected(context.Background(), 2)
	require.NoError(t, err)

	peer1 := net.Hosts()[0]
	peer2 := net.Hosts()[1]

	gossub1, err := pubsub.NewGossipSub(context.Background(), peer1,
		pubsub.WithMessageSignaturePolicy(pubsub.StrictNoSign))
	require.NoError(t, err)
	gossub2, err := pubsub.NewGossipSub(context.Background(), peer2,
		pubsub.WithMessageSignaturePolicy(pubsub.StrictNoSign))
	require.NoError(t, err)

	topic1, err := gossub1.Join("test-header")
	require.NoError(t, err)
	topic2, err := gossub2.Join("test-header")
	require.NoError(t, err)

	sub1, err := topic1.Subscribe()
	require.NoError(t, err)
	sub2, err := topic2.Subscribe()
	require.NoError(t, err)

	// sleep to give it time to initialize
	time.Sleep(time.Millisecond * 100)

	err = topic1.Publish(context.Background(), []byte("hello"))
	require.NoError(t, err)

	// read next message from topic
	msg, err := sub2.Next(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), msg)

	msg, err = sub1.Next(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), msg)
}
