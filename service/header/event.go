package header

import "context"

// coreListener listens for new header events from Core
type coreListener struct {
	sub         *coreSubscription
	sync        *Syncer
	broadcaster Broadcaster
}

// NewCoreListener returns a new Header coreListener.
func NewCoreListener(sub *coreSubscription, sync *Syncer, broadcaster Broadcaster) *coreListener {
	return &coreListener{
		sub:         sub,
		sync:        sync,
		broadcaster: broadcaster,
	}
}

// listen begins listening for new header events from Core once node is
// done syncing the chain to head.
func (cl *coreListener) listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-cl.sync.done:
			header, err := cl.sub.NextHeader(context.Background())
			if err != nil {
				// TODO @renaynay: how to handle this err?
			}
			err = cl.broadcaster.Broadcast(context.Background(), header)
			if err != nil {
				log.Errorw("broadcast new ExtendedHeader", "err", err)
				return
			}
		}
	}
}
