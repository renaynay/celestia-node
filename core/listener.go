package core

import (
	"context"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/tendermint/tendermint/types"

	libhead "github.com/celestiaorg/go-header"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/share/eds"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexsub"
)

// Listener is responsible for listening to Core for
// new block events and converting new Core blocks into
// the main data structure used in the Celestia DA network:
// `ExtendedHeader`. After digesting the Core block, extending
// it, and generating the `ExtendedHeader`, the Listener
// broadcasts the new `ExtendedHeader` to the header-sub gossipsub
// network.
type Listener struct {
	fetcher *BlockFetcher

	construct header.ConstructFn
	store     *eds.Store

	headerBroadcaster libhead.Broadcaster[*header.ExtendedHeader]
	hashBroadcaster   shrexsub.BroadcastFn

	cancel context.CancelFunc
}

func NewListener(
	bcast libhead.Broadcaster[*header.ExtendedHeader],
	fetcher *BlockFetcher,
	hashBroadcaster shrexsub.BroadcastFn,
	construct header.ConstructFn,
	store *eds.Store,
) *Listener {
	return &Listener{
		fetcher:           fetcher,
		headerBroadcaster: bcast,
		hashBroadcaster:   hashBroadcaster,
		construct:         construct,
		store:             store,
	}
}

// Start kicks off the Listener listener loop.
func (cl *Listener) Start(context.Context) error {
	if cl.cancel != nil {
		return fmt.Errorf("listener: already started")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cl.cancel = cancel

	sub, err := cl.fetcher.SubscribeNewBlockEvent(ctx)
	if err != nil {
		return err
	}
	go cl.listen(ctx, sub)
	return nil
}

// Stop stops the listener loop.
func (cl *Listener) Stop(ctx context.Context) error {
	cl.cancel()
	cl.cancel = nil
	return cl.fetcher.UnsubscribeNewBlockEvent(ctx)
}

// listen kicks off a loop, listening for new block events from Core,
// generating ExtendedHeaders and broadcasting them to the header-sub
// gossipsub network.
func (cl *Listener) listen(ctx context.Context, sub <-chan types.EventDataSignedBlock) {
	defer log.Info("listener: listening stopped")
	for {
		select {
		case b, ok := <-sub:
			if !ok {
				return
			}
			log.Debugw("listener: new block from core", "height", b.Header.Height)

			syncing, err := cl.fetcher.IsSyncing(ctx)
			if err != nil {
				log.Errorw("listener: getting sync state", "err", err)
				return
			}

			// extend block data
			eds, err := extendBlock(b.Data)
			if err != nil {
				log.Errorw("listener: extending block data", "err", err)
				return
			}

			// generate extended header
			eh, err := cl.construct(ctx, &b.Header, &b.Commit, &b.ValidatorSet, eds)
			if err != nil {
				log.Errorw("listener: making extended header", "err", err)
				return
			}

			// attempt to store block data if not empty
			err = storeEDS(ctx, b.Header.DataHash.Bytes(), eds, cl.store)
			if err != nil {
				log.Errorw("listener: storing EDS", "err", err)
				return
			}

			// notify network of new EDS hash only if core is already synced
			if !syncing {
				err = cl.hashBroadcaster(ctx, b.Header.DataHash.Bytes())
				if err != nil {
					log.Errorw("listener: broadcasting data hash",
						"height", b.Header.Height,
						"hash", b.Header.Hash(), "err", err) //TODO: hash or datahash?
				}
			}

			// broadcast new ExtendedHeader, but if core is still syncing, notify only local subscribers
			err = cl.headerBroadcaster.Broadcast(ctx, eh, pubsub.WithLocalPublication(syncing))
			if err != nil {
				log.Errorw("listener: broadcasting next header",
					"height", b.Header.Height,
					"err", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
