package relay

import (
	"context"

	"github.com/tendermint/tendermint/types"

	libhead "github.com/celestiaorg/go-header"
	"github.com/celestiaorg/rsmt2d"

	"github.com/celestiaorg/celestia-node/core"
	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/share/eds"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexsub"
)

// DARelayer TODO @renaynay
type DARelayer struct {
	construct header.ConstructFn

	fetcher *core.BlockFetcher

	headerBroadcaster libhead.Broadcaster[*header.ExtendedHeader]
	hashBroadcaster   shrexsub.BroadcastFn

	// stores the EDS and makes it available to the DA network
	edsStore *eds.Store
}

func NewDARelayer(
	construct header.ConstructFn,
	fetcher *core.BlockFetcher,
	headerBroadCaster libhead.Broadcaster[*header.ExtendedHeader],
	hashBroadcaster shrexsub.BroadcastFn,
	edsStore *eds.Store,
) *DARelayer {
	return &DARelayer{
		construct:         construct,
		fetcher:           fetcher,
		headerBroadcaster: headerBroadCaster,
		hashBroadcaster:   hashBroadcaster,
		edsStore:          edsStore,
	}
}

func (b *DARelayer) BroadcastAndStore(
	ctx context.Context,
	rawHeader *types.Header,
	eds *rsmt2d.ExtendedDataSquare,
) {
	// fetch commit and valset
	commit, err := b.fetcher.Commit(ctx, &rawHeader.Height)
	if err != nil {
		// TODO @renaynay: DO SOMETHING
	}
	vals, err := b.fetcher.ValidatorSet(ctx, &rawHeader.Height)
	if err != nil {
		// TODO @renaynay: DO SOMETHING
	}

	eh, err := b.construct(rawHeader, commit, vals, eds)
	if err != nil {
		// TODO @renaynay: DO SOMETHING
	}
	if err := eh.Validate(); err != nil {
		// TODO @renaynay: DO SOMETHING
	}

	// broadcast the header
	err = b.headerBroadcaster.Broadcast(ctx, eh)
	if err != nil {
		// TODO @renaynay: DO SOMETHING
	}
	// TODO eventually we need to bring back caching of proofs
	// store the EDS
	err = b.edsStore.Put(ctx, eh.DAH.Hash(), eds)
	if err != nil {
		// TODO @renaynay: DO SOMETHING
	}
	// broadcast the hash
	notif := shrexsub.Notification{
		DataHash: eh.DAH.Hash(),
		Height:   eh.Height(),
	}
	if err := b.hashBroadcaster(ctx, notif); err != nil {
		// TODO @renaynay: DO SOMETHING
	}
}

// TODO create a retry func that catches and retries for ctx.Timeout
