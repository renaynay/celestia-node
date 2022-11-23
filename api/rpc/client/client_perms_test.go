package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-node/api/rpc/permissions"
)

// TODO @renaynay: bad test
func TestClientPermissions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	dummy := permissions.DummySecret()

	token, err := permissions.NewTokenWithPerms(dummy, permissions.DefaultPerms)
	require.NoError(t, err)

	cli, err := NewClientWithPerms(ctx, "http://localhost:26658", string(token))
	require.NoError(t, err)

	stats, err := cli.DAS.SamplingStats(ctx)
	require.NoError(t, err)

	t.Log(stats)
}
