package header

import "context"

// coreListener listens for new header events from Core
type coreListener struct {
	sub         *coreSubscription
	sync        *Syncer
	broadcaster Broadcaster
}

// NewCoreListener returns a new Header coreListener.
func NewCoreListener(sub *coreSubscription , serv *Service) *coreListener {
	return &coreListener{
		sub:         sub,
		sync:        serv.syncer,
		broadcaster: serv,
	}
}

// listen begins listening for new header events from Core once node is
// done syncing the chain to head.
func (cl *coreListener) listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debugw("closing out core header listener")
			return
		case <-cl.sync.done:
			header, err := cl.sub.NextHeader(context.Background())
			if err != nil {
				log.Errorw("reading next header from subscription", "err", err)
				return
			}
			err = cl.broadcaster.Broadcast(context.Background(), header)
			if err != nil {
				log.Errorw("broadcasting new ExtendedHeader", "height", header.Height, "err", err)
				return
			}
		}
	}
}
