package rpc

import (
	"context"
	"testing"

	"github.com/tendermint/tendermint/abci/example/kvstore"
	"github.com/tendermint/tendermint/node"
	rpctest "github.com/tendermint/tendermint/rpc/test"
)

func TestNewClient(t *testing.T) {
	_, backgroundNode := newClient(t)
	backgroundNode.Stop()
}

func TestClient_GetStatus(t *testing.T) {
	client, backgroundNode := newClient(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		backgroundNode.Stop()
		cancel()
	}()

	status, err := client.GetStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(status.NodeInfo)
}

func TestClient_GetBlock(t *testing.T) {
	client, backgroundNode := newClient(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		backgroundNode.Stop()
		cancel()
	}()

	// make 3 blocks
	if err := client.Start(); err != nil {
		t.Fatal(err)
	}
	eventChan, err := client.StartBlockSubscription(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		<-eventChan
	}
	if err := client.StopBlockSubscription(ctx); err != nil {
		t.Fatal(err)
	}

	height := int64(2)
	block, err := client.GetBlock(ctx, &height)
	if err != nil {
		t.Fatal(err)
	}

	if block.Block.Height != height {
		t.Fatalf("mismatched block heights: expected %v, got %v", height, block.Block.Height)
	}
}

func TestClient_StartBlockSubscription(t *testing.T) {
	client, backgroundNode := newClient(t)
	if err := client.Start(); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	eventChan, err := client.StartBlockSubscription(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		client.StopBlockSubscription(ctx)
		backgroundNode.Stop()
		cancel()
	}()

	for i := 0; i < 10; i++ {
		event := <-eventChan
		t.Log("NEW BLOCK: ", event.Data)
	}
}

func newClient(t *testing.T) (*Client, *node.Node) {
	backgroundNode := startTendermint()

	endpoint := backgroundNode.Config().RPC.ListenAddress
	// separate the protocol from the endpoint
	protocol, ip := endpoint[:3], endpoint[6:]

	client, err := NewClient(protocol, ip)
	if err != nil {
		t.Fatal(err)
	}
	return client, backgroundNode
}

func startTendermint() *node.Node {
	app := kvstore.NewApplication()
	app.RetainBlocks = 10
	return rpctest.StartTendermint(app, rpctest.SuppressStdout)
}
