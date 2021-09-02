package block

import (
	"context"
	"github.com/celestiaorg/celestia-node/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/celestiaorg/celestia-node/node"
	"github.com/celestiaorg/celestia-node/node/config"
)

func TestService_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func(){
		cancel()
	})

	stack, err := node.NewFull(ctx, &config.Config{

	})
	if err != nil {
		t.Fatal(err)
	}
}

func startCoreBackground() {
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

}
