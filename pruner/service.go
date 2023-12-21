package pruner

import (
	"context"
	"fmt"
	"github.com/ipfs/go-datastore"
	"time"

	hdr "github.com/celestiaorg/go-header"

	"github.com/celestiaorg/celestia-node/header"
)

// Service handles the pruning routine for the node using the
// prune Pruner.
type Service struct {
	pruner Pruner
	window AvailabilityWindow

	getter hdr.Getter[*header.ExtendedHeader]

	checkpoint *checkpoint
	blockTime  time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	doneCh chan struct{}

	params Params
}

func NewService(
	p Pruner,
	window AvailabilityWindow,
	getter hdr.Getter[*header.ExtendedHeader],
	ds datastore.Datastore,
	blockTime time.Duration,
	opts ...Option,
) *Service {
	params := DefaultParams()
	for _, opt := range opts {
		opt(&params)
	}

	return &Service{
		pruner:     p,
		window:     window,
		getter:     getter,
		checkpoint: newCheckpoint(ds),
		blockTime:  blockTime,
		doneCh:     make(chan struct{}),
		params:     params,
	}
}

func (s *Service) Start(context.Context) error {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	go s.prune()
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	s.cancel()

	select {
	case <-s.doneCh:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("pruner unable to exit within context deadline")
	}
}

func (s *Service) prune() {
	ticker := time.NewTicker(s.params.gcCycle)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			close(s.doneCh)
			return
		case <-ticker.C:
			headers, err := s.findPruneableHeaders(s.ctx)
			if err != nil {
				// TODO @renaynay: record + report errors properly
				continue
			}
			// TODO @renaynay: make deadline a param ? / configurable?
			pruneCtx, cancel := context.WithDeadline(s.ctx, time.Now().Add(time.Minute))
			err = s.pruner.Prune(pruneCtx, headers...)
			cancel()
			if err != nil {
				// TODO @renaynay: record + report errors properly
				continue
			}

			err = s.updateCheckpoint(s.ctx, headers[len(headers)-1])
			if err != nil {
				// TODO @renaynay: record + report errors properly
			}
		}
	}
}
