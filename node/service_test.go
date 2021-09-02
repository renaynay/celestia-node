package node

import (
	"context"
	"github.com/celestiaorg/celestia-node/node/config"
	"github.com/celestiaorg/celestia-node/rpc"
	"testing"
)

// TODO test BlockService -- how to reach into node and see if BlockService is there / started?
func TestService_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func(){
		cancel()
	})

	coreNode := startTendermint()
	endpoint := coreNode.Config().RPC.ListenAddress
	protocol, ip := endpoint[:3], endpoint[6:]

	full, err := NewFull(ctx, &config.Config{
		RPCConfig: &rpc.Config{
			Protocol:   protocol,
			RemoteAddr: ip,
		},
	})
	if err := full.Start(ctx); err != nil {
		t.Fatal(err)
	}
	// get 3 blocks
	if err := full.RPCClient.Start(); err != nil {
		t.Fatal(err)
	}
	blockChan, err := full.RPCClient.StartBlockSubscription(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		<- blockChan
	}
	if err := full.RPCClient.StopBlockSubscription(ctx); err != nil {
		t.Fatal(err)
	}
	if err := full.Stop(ctx); err != nil {
		t.Fatal(err)
	}
}

