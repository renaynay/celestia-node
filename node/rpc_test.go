package node

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/service/header"
	"github.com/celestiaorg/celestia-node/service/rpc"
)

func TestNamespacedSharesRequest(t *testing.T) {
	nd := setupNodeWithModifiedRPC(t)
	// create request for header at height 2
	height := uint64(2)
	endpoint := fmt.Sprintf("http://127.0.0.1:%s/namespaced_shares/0000000000000001/height/%d",
		nd.RPCServer.ListenAddr()[5:], height)
	resp, err := http.Get(endpoint)
	require.NoError(t, err)
	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()
	// check to make sure request was successfully completed
	require.True(t, resp.StatusCode == http.StatusOK)
	// decode resp
	namespacedShares := new(rpc.NamespacedSharesResponse)
	err = json.NewDecoder(resp.Body).Decode(namespacedShares)
	require.Equal(t, height, namespacedShares.Height)
}

func TestHeaderRequest(t *testing.T) {
	nd := setupNodeWithModifiedRPC(t)
	// create request for header at height 2
	endpoint := fmt.Sprintf("http://127.0.0.1:%s/header/2", nd.RPCServer.ListenAddr()[5:])
	resp, err := http.Get(endpoint)
	require.NoError(t, err)
	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()
	// check to make sure request was successfully completed
	require.True(t, resp.StatusCode == http.StatusOK)
}

func setupNodeWithModifiedRPC(t *testing.T) *Node {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	// create test node with a dummy header service, manually add a dummy header
	// service and register it with rpc handler/server
	hServ := setupHeaderService(t)
	// create overrides
	overrideHeaderServ := func(sets *settings) {
		sets.opts = append(sets.opts, fx.Replace(hServ))
	}
	overrideRPCHandler := func(sets *settings) {
		sets.opts = append(sets.opts, fx.Replace(func(srv *rpc.Server) *rpc.Handler {
			handler := rpc.NewHandler(nil, nil, hServ)
			handler.RegisterEndpoints(srv)
			return handler
		}))
	}
	nd := TestNode(t, Light, overrideHeaderServ, overrideRPCHandler)
	// start node
	err := nd.Start(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = nd.Stop(ctx)
		require.NoError(t, err)
	})
	return nd
}

func setupHeaderService(t *testing.T) *header.Service {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	suite := header.NewTestSuite(t, 1)
	head := suite.Head()
	// create stores
	remoteStore := header.NewTestStore(ctx, t, head)
	localStore := header.NewTestStore(ctx, t, head)
	_, err := localStore.Append(ctx, suite.GenExtendedHeaders(5)...)
	require.NoError(t, err)
	// create syncer
	syncer := header.NewSyncer(header.NewLocalExchange(remoteStore), localStore, &header.DummySubscriber{})

	return header.NewHeaderService(syncer, nil, nil, nil)
}
