package das

import (
	"context"
	"sync"
	"time"

	libhead "github.com/celestiaorg/go-header"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/share/shwap/p2p/shrex/shrexsub"
)

// samplingCoordinator runs and coordinates sampling workers and updates current sampling state
type samplingCoordinator struct {
	concurrencyLimit int
	samplingTimeout  time.Duration

	getter      libhead.Getter[*header.ExtendedHeader]
	sampleFn    sampleFn
	broadcastFn shrexsub.BroadcastFn

	state coordinatorState

	// resultCh fans-in sampling results from worker to coordinator
	resultCh chan result
	// updHeadCh signals to update network head header height
	updHeadCh chan *header.ExtendedHeader
	// waitCh signals to block coordinator for external access to state
	waitCh chan *sync.WaitGroup
	// deleteCh signals that headers at or below a height are being deleted
	deleteCh chan uint64

	// cancelFuncs stores cancel functions for each in-progress job
	cancelFuncs map[int]context.CancelFunc

	workersWg sync.WaitGroup
	metrics   *metrics
	done
}

// result will carry errors to coordinator after worker finishes the job
type result struct {
	job
	failed map[uint64]int
	err    error
}

func newSamplingCoordinator(
	params Parameters,
	getter libhead.Getter[*header.ExtendedHeader],
	sample sampleFn,
	broadcast shrexsub.BroadcastFn,
) *samplingCoordinator {
	return &samplingCoordinator{
		concurrencyLimit: params.ConcurrencyLimit,
		samplingTimeout:  params.SampleTimeout,
		getter:           getter,
		sampleFn:         sample,
		broadcastFn:      broadcast,
		state:            newCoordinatorState(params),
		resultCh:         make(chan result),
		updHeadCh:        make(chan *header.ExtendedHeader),
		waitCh:           make(chan *sync.WaitGroup),
		deleteCh:         make(chan uint64, 16), // buffered to avoid blocking OnDelete callback
		cancelFuncs:      make(map[int]context.CancelFunc),
		done:             newDone("sampling coordinator"),
	}
}

func (sc *samplingCoordinator) run(ctx context.Context, cp checkpoint) {
	sc.state.resumeFromCheckpoint(cp)

	// resume workers
	for _, wk := range cp.Workers {
		sc.runWorker(ctx, sc.state.newJob(wk.JobType, wk.From, wk.To))
	}

	for {
		for !sc.concurrencyLimitReached() {
			next, found := sc.state.nextJob()
			if !found {
				break
			}
			sc.runWorker(ctx, next)
		}

		select {
		case head := <-sc.updHeadCh:
			if sc.state.isNewHead(head.Height()) {
				if !sc.recentJobsLimitReached() {
					sc.runWorker(ctx, sc.state.recentJob(head))
				}
				sc.state.updateHead(head.Height())
				// run worker without concurrency limit restrictions to reduced delay
				sc.metrics.observeNewHead(ctx)
			}
		case res := <-sc.resultCh:
			sc.state.handleResult(res)
			// clean up cancel function for completed job
			delete(sc.cancelFuncs, res.id)
		case wg := <-sc.waitCh:
			wg.Wait()
		case height := <-sc.deleteCh:
			sc.handleHeaderDelete(height)
		case <-ctx.Done():
			sc.workersWg.Wait()
			sc.indicateDone()
			return
		}
	}
}

// runWorker runs job in separate worker go-routine
func (sc *samplingCoordinator) runWorker(ctx context.Context, j job) {
	w := newWorker(j, sc.getter, sc.sampleFn, sc.broadcastFn, sc.metrics)
	sc.state.putInProgress(j.id, w.getState)

	// create a cancellable context for this worker
	workerCtx, cancel := context.WithCancel(ctx)
	sc.cancelFuncs[j.id] = cancel

	// launch worker go-routine
	sc.workersWg.Add(1)
	go func() {
		defer sc.workersWg.Done()
		w.run(workerCtx, sc.samplingTimeout, sc.resultCh)
	}()
}

// listen notifies the coordinator about a new network head received via subscription.
func (sc *samplingCoordinator) listen(ctx context.Context, h *header.ExtendedHeader) {
	select {
	case sc.updHeadCh <- h:
	case <-ctx.Done():
	}
}

// onHeaderDelete notifies the coordinator about a header deletion.
func (sc *samplingCoordinator) onHeaderDelete(ctx context.Context, height uint64) error {
	select {
	case sc.deleteCh <- height:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// stats pauses the coordinator to get stats in a concurrently safe manner
func (sc *samplingCoordinator) stats(ctx context.Context) (SamplingStats, error) {
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Done()

	select {
	case sc.waitCh <- &wg:
	case <-ctx.Done():
		return SamplingStats{}, ctx.Err()
	}

	return sc.state.unsafeStats(), nil
}

func (sc *samplingCoordinator) getCheckpoint(ctx context.Context) (checkpoint, error) {
	stats, err := sc.stats(ctx)
	if err != nil {
		return checkpoint{}, err
	}
	return newCheckpoint(stats), nil
}

// concurrencyLimitReached indicates whether concurrencyLimit has been reached
func (sc *samplingCoordinator) concurrencyLimitReached() bool {
	return len(sc.state.inProgress) >= sc.concurrencyLimit
}

// recentJobsLimitReached indicates whether concurrency limit for recent jobs has been reached
func (sc *samplingCoordinator) recentJobsLimitReached() bool {
	return len(sc.state.inProgress) >= 2*sc.concurrencyLimit
}

// handleHeaderDelete handles the deletion of headers at or below the given height.
// It cancels in-flight operations and cleans up state for heights at or below the given height.
func (sc *samplingCoordinator) handleHeaderDelete(height uint64) {
	log.Debugw("handling header delete", "height", height)

	// Cancel all in-progress jobs that are sampling heights at or below the deleted height
	// Note: We cancel jobs where ALL heights being sampled are at or below the deleted height.
	// Jobs that span across the deleted height will continue but will handle missing headers gracefully.
	for jobID, cancel := range sc.cancelFuncs {
		getState, ok := sc.state.inProgress[jobID]
		if !ok {
			continue
		}
		state := getState()
		// Check if this job is working ONLY on heights at or below the deleted height
		if state.to <= height {
			log.Debugw("cancelling job due to header delete",
				"job_id", jobID,
				"job_type", state.jobType,
				"from", state.from,
				"to", state.to,
				"deleted_height", height)
			cancel()
			// Remove from inProgress immediately since the worker won't send a result
			delete(sc.state.inProgress, jobID)
			delete(sc.cancelFuncs, jobID)
		}
	}

	// Clean up state maps
	sc.state.cleanupDeletedHeights(height)
}
