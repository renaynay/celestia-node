package das

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"

	"github.com/celestiaorg/celestia-node/service/header"
	"github.com/celestiaorg/celestia-node/service/share"
)

var log = logging.Logger("das")

// DASer continuously validates availability of data committed to headers.
type DASer struct {
	da   share.Availability
	hsub header.Subscriber

	// getter allows the DASer to fetch an ExtendedHeader (EH) at a certain height
	// and blocks until the EH has been processed by the header store.
	getter HeaderGetter
	// checkpoint store
	cstore datastore.Datastore

	cancel context.CancelFunc

	jobsCh chan *catchUpJob

	sampleDn  chan struct{} // done signal for sample loop
	catchUpDn chan struct{} // done signal for catchUp loop
}

// NewDASer creates a new DASer.
func NewDASer(
	da share.Availability,
	hsub header.Subscriber,
	getter HeaderGetter,
	cstore datastore.Datastore,
) *DASer {
	wrappedDS := wrapCheckpointStore(cstore)
	return &DASer{
		da:     da,
		hsub:   hsub,
		getter: getter,
		cstore: wrappedDS,
		jobsCh: make(chan *catchUpJob, 4), // TODO @renaynay: is this a reasonable buffer?
	}
}

// Start initiates subscription for new ExtendedHeaders and spawns a sampling routine.
func (d *DASer) Start(context.Context) error {
	if d.cancel != nil {
		return fmt.Errorf("da: DASer already started")
	}

	sub, err := d.hsub.Subscribe()
	if err != nil {
		return err
	}

	// load latest DASed checkpoint
	checkpoint, err := loadCheckpoint(d.cstore)
	if err != nil {
		return err
	}
	log.Infow("loaded checkpoint", "height", checkpoint)

	dasCtx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel
	d.sampleDn, d.catchUpDn = make(chan struct{}), make(chan struct{})

	// kick off catch-up routine manager
	go d.catchUpManager(dasCtx, checkpoint)
	// kick off sampling routine for recently received headers
	go d.sample(dasCtx, sub, checkpoint)
	return nil
}

// Stop stops sampling.
func (d *DASer) Stop(ctx context.Context) error {
	d.cancel()
	// wait for both sampling routines to exit
	for i := 0; i < 2; i++ {
		select {
		case <-d.catchUpDn:
			d.catchUpDn = nil
		case <-d.sampleDn:
			d.sampleDn = nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	d.cancel = nil
	return nil
}

// sample validates availability for each Header received from header subscription.
func (d *DASer) sample(ctx context.Context, sub header.Subscription, checkpoint int64) {
	defer func() {
		sub.Cancel()
		close(d.sampleDn)
	}()

	for {
		h, err := sub.NextHeader(ctx)
		if err != nil {
			if err == context.Canceled {
				return
			}

			log.Errorw("failed to get next header", "err", err)
			continue
		}

		// If the next header coming through gossipsub is not adjacent
		// to our last DASed header, kick off routine to DAS all headers
		// between last DASed header and h. This situation could occur
		// either on start or due to network latency/disconnection.
		if h.Height > checkpoint+1 {
			// DAS headers between last DASed height up to the current
			// header
			job := &catchUpJob{
				from: checkpoint,
				to:   h.Height - 1,
			}
			select {
			case <-ctx.Done():
				return
			case d.jobsCh <- job:
			}
		}

		startTime := time.Now()

		err = d.da.SharesAvailable(ctx, h.DAH)
		if err != nil {
			if err == context.Canceled {
				return
			}
			log.Errorw("sampling failed", "height", h.Height, "hash", h.Hash(),
				"square width", len(h.DAH.RowsRoots), "data root", h.DAH.Hash(), "err", err)
			// continue sampling
		}

		sampleTime := time.Since(startTime)
		log.Infow("sampling successful", "height", h.Height, "hash", h.Hash(),
			"square width", len(h.DAH.RowsRoots), "finished (s)", sampleTime.Seconds())

		checkpoint = h.Height
	}
}

// catchUpJob represents a catch-up job. (from:to]
type catchUpJob struct {
	from, to int64
}

// catchUpResult contains the result of a catch-up routine for
// a single `catchUpJob`.
type catchUpResult struct {
	checkpoint int64 // last successfully sampled height
	err        error // error that occurred during catch-up routine
}

// catchUpManager manages catch-up jobs, performing them one at a time, exiting
// only once context is canceled and storing latest DASed checkpoint to disk.
func (d *DASer) catchUpManager(ctx context.Context, checkpoint int64) {
	defer func() {
		// store latest DASed checkpoint to disk here to ensure that if DASer is not yet
		// fully caught up to network head, it will resume DASing from this checkpoint
		// up to current network head
		// TODO @renaynay: Implement Share Cache #180 to ensure no duplicate DASing over same
		//  header
		if err := storeCheckpoint(d.cstore, checkpoint); err != nil {
			log.Errorw("storing checkpoint to disk", "height", checkpoint, "err", err)
		}
		// signal that all catchUp routines have finished
		close(d.catchUpDn)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case job := <-d.jobsCh:
			// perform catchUp routine
			result := d.catchUp(ctx, job)
			// update checkpoint before checking error as the checkpoint
			// value from result represents the last successfully DASed height
			// of the routine's success
			checkpoint = result.checkpoint
			// exit routine if a catch-up job was unsuccessful
			if result.err != nil {
				log.Errorw("catch-up routine failed", "attempted range (from, to)", job.from,
					job.to, "last successfully sampled height", result.checkpoint)
				return
			}
		}
	}
}

// catchUp starts a sampling routine for headers starting at the next header
// after the `from` height and exits the loop once `to` is reached. (from:to]
func (d *DASer) catchUp(ctx context.Context, job *catchUpJob) *catchUpResult {
	routineStartTime := time.Now()
	log.Infow("sampling past headers", "from", job.from, "to", job.to)

	// start sampling from height at checkpoint+1 since the
	// checkpoint height is DASed by broader sample routine
	for height := job.from + 1; height <= job.to; height++ {
		h, err := d.getter.GetByHeight(ctx, uint64(height))
		if err != nil {
			if err == context.Canceled {
				return &catchUpResult{
					checkpoint: height - 1, // previous height is the last successfully sampled height
					err:        nil,        // report error as nil as routine was ordered to stop
				}
			}

			log.Errorw("failed to get next header", "height", height, "err", err)
			return &catchUpResult{
				checkpoint: height - 1, // previous height is the last successfully sampled height
				err:        err,
			}
		}

		startTime := time.Now()

		err = d.da.SharesAvailable(ctx, h.DAH)
		if err != nil {
			if err == context.Canceled {
				return &catchUpResult{
					checkpoint: height - 1, // previous height is the last successfully sampled height
					err:        nil,        // report error as nil as routine was ordered to stop
				}

			}
			log.Errorw("sampling failed", "height", h.Height, "hash", h.Hash(),
				"square width", len(h.DAH.RowsRoots), "data root", h.DAH.Hash(), "err", err)
			// continue sampling
		}

		sampleTime := time.Since(startTime)
		log.Debugw("sampled past header", "height", h.Height, "hash", h.Hash(),
			"square width", len(h.DAH.RowsRoots), "finished (s)", sampleTime.Seconds())
	}

	log.Infow("successfully caught up", "from", job.from,
		"to", job.to, "finished (s)", time.Since(routineStartTime))
	// report successful result
	return &catchUpResult{
		checkpoint: job.to,
		err:        nil,
	}
}
