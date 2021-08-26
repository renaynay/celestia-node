package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-node/node/config"
	"github.com/celestiaorg/celestia-node/rpc"
	"github.com/tendermint/tendermint/abci/example/kvstore"
	tend_node "github.com/tendermint/tendermint/node"
	rpctest "github.com/tendermint/tendermint/rpc/test"
)

func TestNewFull(t *testing.T) {
	startCtx, cancel := context.WithCancel(context.Background())

	tendNode := startTendermint()
	defer func() {
		tendNode.Stop()
		cancel()
	}()
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

	stopCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = node.Stop(stopCtx)
	assert.NoError(t, err)
}

func startTendermint() *tend_node.Node {
	app := kvstore.NewApplication()
	app.RetainBlocks = 10
	return rpctest.StartTendermint(app, rpctest.SuppressStdout)
}
