package header

import (
	"context"
	"fmt"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"testing"
	"time"

	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/require"
)

func TestCoreListener(t *testing.T) {
	// create mocknet with 2 peers, one set up to listen on topic
	net, err := mocknet.FullMeshConnected(context.Background(), 2)
	require.NoError(t, err)

	host, peer := net.Hosts()[0], net.Hosts()[1]
	fmt.Println(peer)

	hostPS, err := pubsub.NewGossipSub(context.Background(), host,
		pubsub.WithMessageSignaturePolicy(pubsub.StrictNoSign))
	require.NoError(t, err)
//	peerPS, err := pubsub.NewGossipSub(context.Background(), peer,
//		pubsub.WithMessageSignaturePolicy(pubsub.StrictNoSign))

	// set up header service with coreEnabled
	store := createStore(t, 10)
	fetcher := createMockCoreFetcher()
	coreEx := NewCoreExchange(fetcher, nil)

	head, err := store.Head(context.Background())
	require.NoError(t, err)

	syncer := NewSyncer(coreEx, store, head.Hash())
	serv := NewHeaderService(syncer, hostPS, true)
	// see if extendedheader gets broadcasted to other side
	err = serv.Start(context.Background())
	require.NoError(t, err)

	time.Sleep(10)

	require.NoError(t, serv.Stop(context.Background()))
}
