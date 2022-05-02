package tests

import (
	"context"
	appnetutil "github.com/celestiaorg/celestia-app/testutil/network"
	"github.com/celestiaorg/celestia-node/node"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/os"
	"testing"
)

func TestState_SubmitPayForData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// set up app + core node w app
	network := appnetutil.New(t, appnetutil.DefaultConfig())
	valAddr := network.Validators[0].AppConfig.GRPC.Address

	cfg := node.DefaultConfig(node.Light)
	store := node.MockStore(t, cfg)

	ks, err := store.Keystore()
	require.NoError(t, err)
	path := ks.Path()
	os.WriteFile(path, []byte(""))

	ln, err := node.New(node.Light, store, node.WithGRPCEndpoint(valAddr))
	require.NoError(t, err)
	err = ln.Start(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = ln.Stop(ctx)
		require.NoError(t, err)
	})
}
