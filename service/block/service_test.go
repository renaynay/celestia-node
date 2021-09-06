package block

import (
	"context"
	"github.com/celestiaorg/celestia-core/abci/example/kvstore"
	"github.com/celestiaorg/celestia-core/node"
	rpctest "github.com/celestiaorg/celestia-core/rpc/test"
	"github.com/celestiaorg/celestia-node/node/rpc"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewBlockService(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	coreNode := startCoreNode()
	//nolint:errcheck
	t.Cleanup(func(){
		coreNode.Stop()
		cancel()
	})

	endpoint := coreNode.Config().RPC.ListenAddress
	// separate the protocol from the endpoint
	protocol, ip := endpoint[:3], endpoint[6:]

	client, err := rpc.NewClient(protocol, ip)
	require.Nil(t, err)

	blockServ := NewBlockService(client)

	errCh := blockServ.Start(ctx)
	for {
		select{
		case err := <- errCh:
			t.Fatal(err)
		}
	}

	err = blockServ.Stop(ctx)
	require.Nil(t, err)
}

func startCoreNode() *node.Node {
	app := kvstore.NewApplication()
	app.RetainBlocks = 10
	return rpctest.StartTendermint(app, rpctest.SuppressStdout)
}
