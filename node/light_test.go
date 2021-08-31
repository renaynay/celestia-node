package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-node/node/config"
)

func TestNewLight(t *testing.T) {
	startCtx, startCtxCancel := context.WithCancel(context.Background())

	nd, err := NewLight(startCtx, &config.Config{})
	assert.NoError(t, err)
	require.NotNil(t, nd)

	stopCtx, stopCtxCancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		startCtxCancel()
		stopCtxCancel()
	})

	err = nd.Stop(stopCtx)
	assert.NoError(t, err)
}
