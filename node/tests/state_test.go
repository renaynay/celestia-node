package tests

import (
	"context"
	"github.com/celestiaorg/celestia-node/node/tests/swamp"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"testing"
)

func TestState_SubmitPayForData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	sw := &swamp.Swamp{
		Network: mocknet.New(ctx),
	}
	app := sw.NewAppInstance(t)
	app
}
