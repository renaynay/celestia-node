package relay

import (
	"context"

	"github.com/tendermint/tendermint/types"

	libhead "github.com/celestiaorg/go-header"
	"github.com/celestiaorg/rsmt2d"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/share/eds"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexsub"
)

// DARelayer TODO @renaynay
type DARelayer struct {
	construct header.ConstructFn

	headerBroadcaster libhead.Broadcaster[*header.ExtendedHeader]
	hashBroadcaster   shrexsub.BroadcastFn

	// stores the EDS and makes it available to the DA network
	edsStore *eds.Store
}

func NewDARelayer(
	construct header.ConstructFn,
	headerBroadCaster libhead.Broadcaster[*header.ExtendedHeader],
	hashBroadcaster shrexsub.BroadcastFn,
	edsStore *eds.Store,
) *DARelayer {
	return &DARelayer{
		construct:         construct,
		headerBroadcaster: headerBroadCaster,
		hashBroadcaster:   hashBroadcaster,
		edsStore:          edsStore,
	}
}

func (b *DARelayer) BroadcastAndStore(
	ctx context.Context,
	rawHeader *types.Header,
	commit *types.Commit,
	vals *types.ValidatorSet,
	eds *rsmt2d.ExtendedDataSquare,
) error {
	eh, err := b.construct(rawHeader, commit, vals, eds)
	if err != nil {
		return err
	}
	if err := eh.ValidateBasic(); err != nil {
		return err
	}
	// broadcast the header
	err = b.headerBroadcaster.Broadcast(ctx, eh)
	if err != nil {
		return err
	}
	// TODO eventually we need to bring back caching of proofs
	// store the EDS
	err = b.edsStore.Put(ctx, eh.DAH.Hash(), eds)
	if err != nil {
		return err
	}
	// broadcast the hash
	notif := shrexsub.Notification{
		DataHash: eh.DAH.Hash(),
		Height:   eh.Height(),
	}
	return b.hashBroadcaster(ctx, notif)
}

// TODO create a retry func that catches and retries for ctx.Timeout
