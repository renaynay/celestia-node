package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-core/abci/example/kvstore"
	core_node "github.com/celestiaorg/celestia-core/node"
	rpctest "github.com/celestiaorg/celestia-core/rpc/test"
	"github.com/celestiaorg/celestia-node/node/config"
	"github.com/celestiaorg/celestia-node/rpc"
)

func TestNewFull(t *testing.T) {
	startCtx, startCtxCancel := context.WithCancel(context.Background())

	tendNode := startTendermint()
	endpoint := tendNode.Config().RPC.ListenAddress
	protocol, ip := endpoint[:3], endpoint[6:]

	node, err := NewFull(startCtx, &config.Config{
		RPCConfig: &rpc.Config{
			Protocol:   protocol,
			RemoteAddr: ip,
		},
	})
	assert.NoError(t, err)
	require.NotNil(t, node)
	require.NotNil(t, node.RPCClient)

	stopCtx, stopCtxCancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		tendNode.Stop()
		startCtxCancel()
		stopCtxCancel()
	})

	err = node.Stop(stopCtx)
	assert.NoError(t, err)
}

func startTendermint() *core_node.Node {
	app := kvstore.NewApplication()
	app.RetainBlocks = 10
	return rpctest.StartTendermint(app, rpctest.SuppressStdout)
}
