package tests

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	mdutils "github.com/ipfs/go-merkledag/test"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-node/header/headertest"
	"github.com/celestiaorg/celestia-node/nodebuilder"
	"github.com/celestiaorg/celestia-node/nodebuilder/core"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/tests/swamp"
	"github.com/celestiaorg/celestia-node/share/eds/byzantine"
)

/*
Test-Case: Full Node will propagate a fraud proof to the network, once ByzantineError will be received from sampling.
Pre-Requisites:
- CoreClient is started by swamp.
Steps:
1. Create a Bridge Node(BN) with fraudulent extended header at height 10.
2. Start a BN.
3. Create a Full Node(FN) with a connection to BN as a trusted peer.
4. Start a FN.
5. Subscribe to bad encoding fraud proof and wait for it to be received.
6. Check FN has not synced beyond 10.

// TODO
Note: 15 is not available because DASer will be stopped before reaching this height due to receiving a fraud proof.
Another note: this test disables share exchange to speed up test results.
*/
func TestFraudProofBroadcasting(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), swamp.DefaultTestTimeout)
	t.Cleanup(cancel)

	const (
		blocks    = 15
		blockSize = 2
		blockTime = time.Millisecond * 300
	)

	sw := swamp.NewSwamp(t, swamp.WithBlockTime(blockTime))
	fillDn := swamp.FillBlocks(ctx, sw.ClientContext, sw.Accounts, blockSize, blocks)

	// create and start a BN that will create a fraudulent header
	// at height 10
	cfg := nodebuilder.DefaultConfig(node.Bridge)
	cfg.Share.UseShareExchange = false
	bridge := sw.NewNodeWithConfig(node.Bridge, cfg, core.WithHeaderConstructFn(headertest.FraudMaker(t, 10, mdutils.Bserv())))
	err := bridge.Start(ctx)
	require.NoError(t, err)

	// start a FN with BN as a trusted peer
	cfg = nodebuilder.DefaultConfig(node.Full)
	cfg.Share.UseShareExchange = false
	swamp.WithTrustedPeers(t, cfg, bridge)
	store := nodebuilder.MockStore(t, cfg)
	full := sw.NewNodeWithStore(node.Full, store)
	err = full.Start(ctx)
	require.NoError(t, err)

	// subscribe to fraud proof before node starts helps
	// to prevent flakiness when fraud proof is propagating before subscribing on it
	subscr, err := full.FraudServ.Subscribe(ctx, byzantine.BadEncoding)
	require.NoError(t, err)

	select {
	case p := <-subscr:
		require.Equal(t, 10, int(p.Height()))
		t.Log("HERE!")
	case <-ctx.Done():
		t.Fatal("fraud proof was not received in time")
	}
	// This is an obscure way to check if the Syncer was stopped.
	// If we cannot get a height header within a timeframe it means the syncer was stopped
	// FIXME: Eventually, this should be a check on service registry managing and keeping
	//  lifecycles of each Module.
	syncCtx, syncCancel := context.WithTimeout(context.Background(), blockTime)
	t.Cleanup(syncCancel)
	_, err = full.HeaderServ.GetByHeight(syncCtx, blocks)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	// start a new FN connected to the previous FN after the fraud proof has
	// already been propagated to ensure that it can sync the fraud proof properly
	// and stall services.
	cfg = nodebuilder.DefaultConfig(node.Full)
	swamp.WithTrustedPeers(t, cfg, full)
	newFull := sw.NewNodeWithConfig(node.Full, cfg)
	err = newFull.Start(ctx)
	require.NoError(t, err)

	time.Sleep(5 * time.Second)

	proofs, err := newFull.FraudServ.Get(ctx, byzantine.BadEncoding)
	require.NoError(t, err)
	assert.Equal(t, 10, proofs[0].Proof.Height())
	assert.NotNil(t, proofs)

	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	case err := <-fillDn:
		require.NoError(t, err)
	}
}

/*
Test-Case: Light node receives a fraud proof using Fraud Sync
Pre-Requisites:
- CoreClient is started by swamp.
Steps:
1. Create a Bridge Node(BN) with broken extended header at height 10.
2. Start a BN.
3. Create a Full Node(FN) with a connection to BN as a trusted peer.
4. Start a FN.
5. Subscribe to a fraud proof and wait when it will be received.
6. Start LN once a fraud proof is received and verified by FN.
7. Wait until LN will be connected to FN and fetch a fraud proof.
Note: this test disables share exchange to speed up test results.
*/
func TestFraudProofSyncing(t *testing.T) {
	const (
		blocks = 15
		bsize  = 2
		btime  = time.Millisecond * 300
	)
	sw := swamp.NewSwamp(t, swamp.WithBlockTime(btime))
	ctx, cancel := context.WithTimeout(context.Background(), swamp.DefaultTestTimeout)
	t.Cleanup(cancel)

	fillDn := swamp.FillBlocks(ctx, sw.ClientContext, sw.Accounts, bsize, blocks)
	cfg := nodebuilder.DefaultConfig(node.Bridge)
	cfg.Share.UseShareExchange = false
	store := nodebuilder.MockStore(t, cfg)
	bridge := sw.NewNodeWithStore(
		node.Bridge,
		store,
		core.WithHeaderConstructFn(headertest.FraudMaker(t, 10, mdutils.Bserv())),
	)

	err := bridge.Start(ctx)
	require.NoError(t, err)
	addr := host.InfoFromHost(bridge.Host)
	addrs, err := peer.AddrInfoToP2pAddrs(addr)
	require.NoError(t, err)

	fullCfg := nodebuilder.DefaultConfig(node.Full)
	fullCfg.Share.UseShareExchange = false
	fullCfg.Header.TrustedPeers = append(fullCfg.Header.TrustedPeers, addrs[0].String())
	full := sw.NewNodeWithStore(node.Full, nodebuilder.MockStore(t, fullCfg))

	lightCfg := nodebuilder.DefaultConfig(node.Light)
	lightCfg.Header.TrustedPeers = append(lightCfg.Header.TrustedPeers, addrs[0].String())
	ln := sw.NewNodeWithStore(node.Light, nodebuilder.MockStore(t, lightCfg))
	require.NoError(t, full.Start(ctx))

	subsFN, err := full.FraudServ.Subscribe(ctx, byzantine.BadEncoding)
	require.NoError(t, err)

	select {
	case <-subsFN:
	case <-ctx.Done():
		t.Fatal("full node didn't get FP in time")
	}

	// start LN to enforce syncing logic, not the PubSub's broadcasting
	err = ln.Start(ctx)
	require.NoError(t, err)

	// internal subscription for the fraud proof is done in order to ensure that light node
	// receives the BEFP.
	subsLN, err := ln.FraudServ.Subscribe(ctx, byzantine.BadEncoding)
	require.NoError(t, err)

	// ensure that the full and light node are connected to speed up test
	// alternatively, they would discover each other
	err = ln.Host.Connect(ctx, *host.InfoFromHost(full.Host))
	require.NoError(t, err)

	select {
	case <-subsLN:
	case <-ctx.Done():
		t.Fatal("light node didn't get FP in time")
	}
	require.NoError(t, <-fillDn)
}
