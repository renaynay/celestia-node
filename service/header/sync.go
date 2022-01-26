package header

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

// Syncer implements efficient synchronization for headers.
//
// There are two main processes running in Syncer:
// 1. Main syncing loop(s.syncLoop)
//    * Performs syncing from the subjective(local chain view) header up to the latest known trusted header
//    * Syncs by requesting missing headers from Exchange or
//    * By accessing cache of pending and verified headers
// 2. Receives new headers from PubSub subnetwork (s.processIncoming)
//    * Usually, a new header is adjacent to the trusted head and if so, it is simply appended to the local store,
//    incrementing the subjective height and making it the new latest known trusted header.
//    * Or, if it receives a header further in the future,
//      * verifies against the latest known trusted header
//    	* adds the header to pending cache(making it the latest known trusted header)
//      * and triggers syncing loop to catch up to that point.
type Syncer struct {
	sub      Subscriber
	exchange Exchange
	store    Store

	// trusted hash of the header from which syncer starts to sync, a.k.a genesis,
	// which can be any valid past header in the chain we trust
	trusted tmbytes.HexBytes
	// stateLk protects state which represents the current or latest sync
	stateLk sync.RWMutex
	state   SyncState
	// inProgress is set to 1 once syncing commences and
	// is set to 0 once syncing is either finished or
	// not currently in progress
	inProgress uint64
	heightSub  *heightSub
	// signals to start syncing
	triggerSync chan struct{}
	// pending keeps ranges of valid headers received from the network awaiting to be appended to store
	pending ranges
	// cancel cancels syncLoop's context
	cancel context.CancelFunc
}

// NewSyncer creates a new instance of Syncer.
func NewSyncer(exchange Exchange, store Store, sub Subscriber, trusted tmbytes.HexBytes) *Syncer {
	return &Syncer{
		sub:         sub,
		exchange:    exchange,
		store:       store,
		trusted:     trusted,
		heightSub:   newHeightSub(store),
		triggerSync: make(chan struct{}, 1), // should be buffered
	}
}

// Start starts the syncing routine.
func (s *Syncer) Start(ctx context.Context) error {
	if s.cancel != nil {
		return fmt.Errorf("header: Syncer already started")
	}

	err := s.sub.AddValidator(s.processIncoming)
	if err != nil {
		return err
	}

	// TODO(@Wondertan): Ideally, this initialization should be part of Init process
	err = s.initStore(ctx)
	if err != nil {
		log.Error(err)
	}
	s.heightSub.Start()

	ctx, cancel := context.WithCancel(context.Background())
	go s.syncLoop(ctx)
	s.wantSync()
	s.cancel = cancel
	return nil
}

// Stop stops Syncer.
func (s *Syncer) Stop(ctx context.Context) error {
	err := s.WaitSync(ctx)
	s.heightSub.Stop()
	s.cancel()
	s.cancel = nil
	return err
}

// IsSyncing returns the current sync status of the Syncer.
func (s *Syncer) IsSyncing() bool {
	return atomic.LoadUint64(&s.inProgress) == 1
}

// WaitSync blocks until ongoing sync is done.
func (s *Syncer) WaitSync(ctx context.Context) error {
	state := s.State()
	if state.Finished() {
		return nil
	}

	_, err := s.GetByHeight(ctx, state.ToHeight)
	return err
}

// SyncState collects all the information about o sync.
type SyncState struct {
	ID                   uint64 // incrementing ID of a sync
	Height               uint64 // height at the moment when State is requested for a sync
	FromHeight, ToHeight uint64 // the starting and the ending point of a sync
	FromHash, ToHash     tmbytes.HexBytes
	Start, End           time.Time
	Error                error // the error that might happen within a sync
}

// Finished is true whether a sync is done.
func (s SyncState) Finished() bool {
	return s.ToHeight == s.Height
}

// State reports state of current, if in progress, or last sync, if finished.
// Note that throughout the whole Syncer lifetime there might an initial sync and multiple catch-ups.
// All of them are treated as different syncs with different state IDs and other information.
func (s *Syncer) State() SyncState {
	s.stateLk.RLock()
	defer s.stateLk.RUnlock()
	return s.state
}

// GetByHeight gets header by height from Store or, if not found, waits until it is available.
func (s *Syncer) GetByHeight(ctx context.Context, height uint64) (*ExtendedHeader, error) {
	return s.heightSub.GetByHeight(ctx, height)
}

// init initializes if it's empty
func (s *Syncer) initStore(ctx context.Context) error {
	_, err := s.store.Head(ctx)
	switch err {
	case ErrNoHead:
		// if there is no head - request header at trusted hash.
		trusted, err := s.exchange.RequestByHash(ctx, s.trusted)
		if err != nil {
			return fmt.Errorf("header: requesting header at trusted hash during init: %w", err)
		}

		err = s.store.Append(ctx, trusted)
		if err != nil {
			return fmt.Errorf("header: appending header during init: %w", err)
		}
	case nil:
	}

	return nil
}

// trustedHead returns the latest known trusted header that is within the trusting period.
func (s *Syncer) trustedHead(ctx context.Context) (*ExtendedHeader, error) {
	// check pending for trusted header and return it if applicable
	// NOTE: Pending cannot be expired, guaranteed
	pendHead := s.pending.Head()
	if pendHead != nil {
		return pendHead, nil
	}

	sbj, err := s.store.Head(ctx)
	if err != nil {
		return nil, err
	}

	// check if our subjective header is not expired and use it
	if !sbj.IsExpired() {
		return sbj, nil
	}

	// otherwise, request head from a trustedPeer or, in other words, do automatic subjective initialization
	objHead, err := s.exchange.RequestHead(ctx)
	if err != nil {
		return nil, err
	}

	s.pending.Add(objHead)
	return objHead, nil
}

// wantSync will trigger the syncing loop (non-blocking).
func (s *Syncer) wantSync() {
	select {
	case s.triggerSync <- struct{}{}:
	default:
	}
}

// syncLoop controls syncing process.
func (s *Syncer) syncLoop(ctx context.Context) {
	for {
		select {
		case <-s.triggerSync:
			s.sync(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// sync ensures we are synced up to any trusted header.
func (s *Syncer) sync(ctx context.Context) {
	// indicate syncing
	atomic.StoreUint64(&s.inProgress, 1)
	// indicate syncing is stopped
	defer atomic.StoreUint64(&s.inProgress, 0)

	trstHead, err := s.trustedHead(ctx)
	if err != nil {
		log.Errorw("getting trusted head", "err", err)
		return
	}

	s.syncTo(ctx, trstHead)
}

// processIncoming processes new processIncoming Headers, validates them and stores/caches if applicable.
func (s *Syncer) processIncoming(ctx context.Context, maybeHead *ExtendedHeader) pubsub.ValidationResult {
	// 1. Try to append. If header is not adjacent/from future - try it for pending cache below
	err := s.store.Append(ctx, maybeHead)
	switch err {
	case nil:
		// a happy case where we append adjacent header correctly
		s.heightSub.ProvideHeights(ctx, maybeHead)
		return pubsub.ValidationAccept
	case ErrNonAdjacent:
		// not adjacent, so try to cache it after verifying
	default:
		var verErr *VerifyError
		if errors.As(err, &verErr) {
			log.Errorw("invalid header",
				"height", maybeHead.Height,
				"hash", maybeHead.Hash(),
				"reason", verErr.Reason)
			return pubsub.ValidationReject
		}

		log.Errorw("appending header",
			"height", maybeHead.Height,
			"hash", maybeHead.Hash().String(),
			"err", err)
		// might be a storage error or something else, but we can still try to continue processing 'maybeHead'
	}

	// 2. Get known trusted head, so we can verify maybeHead
	trstHead, err := s.trustedHead(ctx)
	if err != nil {
		log.Errorw("getting trusted head", "err", err)
		return pubsub.ValidationIgnore // we don't know if header is invalid so ignore
	}

	// 3. Filter out maybeHead if behind trusted
	if maybeHead.Height <= trstHead.Height {
		log.Warnw("received known header",
			"height", maybeHead.Height,
			"hash", maybeHead.Hash())

		// TODO(@Wondertan): Remove once duplicates are fully fixed
		log.Warnf("Ignore the warn above - there is a known issue with duplicate headers on the network.")
		return pubsub.ValidationIgnore // we don't know if header is invalid so ignore
	}

	// 4. Verify maybeHead against trusted
	err = trstHead.VerifyNonAdjacent(maybeHead)
	var verErr *VerifyError
	if errors.As(err, &verErr) {
		log.Errorw("invalid header",
			"height", maybeHead.Height,
			"hash", maybeHead.Hash(),
			"reason", verErr.Reason)
		return pubsub.ValidationReject
	}

	// 5. Save verified header to pending cache
	// NOTE: Pending cache can't be DOSed as we verify above each header against a trusted one.
	s.pending.Add(maybeHead)
	// and trigger sync to catch-up
	s.wantSync()
	log.Infow("new pending head",
		"height", maybeHead.Height,
		"hash", maybeHead.Hash())
	return pubsub.ValidationAccept
}

// TODO(@Wondertan): Number of headers that can be requested at once. Either make this configurable or,
//  find a proper rationale for constant.
var requestSize uint64 = 512

// syncTo requests headers from locally stored head up to the new head.
func (s *Syncer) syncTo(ctx context.Context, newHead *ExtendedHeader) {
	head, err := s.store.Head(ctx)
	if err != nil {
		log.Errorw("getting head during sync", "err", err)
		return
	}

	if head.Height == newHead.Height {
		return
	}

	log.Infow("syncing headers",
		"from", head.Height,
		"to", newHead.Height)
	err = s.doSync(ctx, head, newHead)
	if err != nil {
		log.Errorw("syncing headers",
			"from", head.Height,
			"to", newHead.Height,
			"err", err)
		return
	}

	log.Infow("synced headers",
		"from", head.Height,
		"to", newHead.Height,
		"took", s.state.End.Sub(s.state.Start))
}

// doSync performs actual syncing updating the internal SyncState
func (s *Syncer) doSync(ctx context.Context, oldHead, newHead *ExtendedHeader) (err error) {
	from, to := uint64(oldHead.Height)+1, uint64(newHead.Height)

	s.stateLk.Lock()
	s.state.ID++
	s.state.FromHeight = from
	s.state.ToHeight = to
	s.state.FromHash = oldHead.Hash()
	s.state.ToHash = newHead.Hash()
	s.state.Start = time.Now()
	s.stateLk.Unlock()

	for from <= to {
		amount := to - from + 1
		if amount > requestSize {
			amount = requestSize
		}

		amount, err = s.processHeaders(ctx, from, amount)
		if err != nil && amount == 0 {
			break
		}

		from += amount
		s.stateLk.Lock()
		s.state.Height = from
		s.stateLk.Unlock()
	}

	s.stateLk.Lock()
	s.state.Height = from - 1 // minus one as we add one to the amount above
	s.state.End = time.Now()
	s.state.Error = err
	s.stateLk.Unlock()
	return err
}

// processHeaders gets and stores the 'amount' number of headers going from the 'start' height.
func (s *Syncer) processHeaders(ctx context.Context, from, amount uint64) (uint64, error) {
	headers, err := s.getHeaders(ctx, from, amount)
	if err != nil {
		return 0, err
	}

	// TODO(@Wondertan): Append should report how many headers were applied
	err = s.store.Append(ctx, headers...)
	if err != nil {
		return 0, err
	}

	s.heightSub.ProvideHeights(ctx, headers...)
	return uint64(len(headers)), nil
}

// getHeaders gets headers from either remote peers or from local cache of headers received by PubSub
func (s *Syncer) getHeaders(ctx context.Context, start, amount uint64) ([]*ExtendedHeader, error) {
	// short-circuit if nothing in pending cache to avoid unnecessary allocation below
	if _, ok := s.pending.FirstRangeWithin(start, start+amount); !ok {
		return s.exchange.RequestHeaders(ctx, start, amount)
	}

	end, out := start+amount, make([]*ExtendedHeader, 0, amount)
	for start < end {
		// if we have some range cached - use it
		if r, ok := s.pending.FirstRangeWithin(start, end); ok {
			// first, request everything between start and found range
			hs, err := s.exchange.RequestHeaders(ctx, start, r.Start-start)
			if err != nil {
				return nil, err
			}
			out = append(out, hs...)
			start += uint64(len(hs))

			// then, apply cached range
			cached := r.Before(end)
			out = append(out, cached...)
			start += uint64(len(cached))

			// repeat, as there might be multiple cache ranges
			continue
		}

		// fetch the leftovers
		hs, err := s.exchange.RequestHeaders(ctx, start, end-start)
		if err != nil {
			// still return what was successfully gotten
			return out, err
		}

		return append(out, hs...), nil
	}

	return out, nil
}

// heightSub provides a way to wait until a specific height becomes available and synced
type heightSub struct {
	// usually we don't attach context to structs, but this is an exception
	// as GetByHeight and ProvideHeights needs to access it
	ctx       context.Context
	cancel    context.CancelFunc
	store     Store
	height    uint64
	provideCh chan []*ExtendedHeader
	reqsCh    chan *heightReq
	reqs      map[uint64][]chan *heightResp
}

type heightReq struct {
	resp   chan *heightResp
	height uint64
}

type heightResp struct {
	header *ExtendedHeader
	err    error
}

func newHeightSub(store Store) *heightSub {
	return &heightSub{
		store:     store,
		provideCh: make(chan []*ExtendedHeader, 4),
		reqsCh:    make(chan *heightReq, 4),
		reqs:      make(map[uint64][]chan *heightResp),
	}
}

func (hs *heightSub) Start() {
	hs.ctx, hs.cancel = context.WithCancel(context.Background())
	go hs.subLoop()
}

func (hs *heightSub) Stop() {
	hs.cancel()
}

func (hs *heightSub) GetByHeight(ctx context.Context, height uint64) (*ExtendedHeader, error) {
	h, err := hs.store.GetByHeight(ctx, height)
	if err != ErrNotFound {
		return h, err
	}

	resp := make(chan *heightResp, 1)
	select {
	case hs.reqsCh <- &heightReq{resp, height}:
		select {
		case resp := <-resp:
			return resp.header, resp.err
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-hs.ctx.Done():
			return nil, hs.ctx.Err()
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-hs.ctx.Done():
		return nil, hs.ctx.Err()
	}
}

func (hs *heightSub) ProvideHeights(ctx context.Context, headers ...*ExtendedHeader) {
	select {
	case hs.provideCh <- headers:
	case <-ctx.Done():
	case <-hs.ctx.Done():
	}
}

func (hs *heightSub) subLoop() {
	for {
		select {
		case req := <-hs.reqsCh:
			if req.height <= hs.height {
				// this is a very rare case which can happen when the following flow happens
				// 1. store.GetByHeight in hs.GetByHeight reports ErrNotFound for the requested height
				// 2. store is appended with range including the requested height and headerSub is provided with them
				// 3. headerSub processes provide before the request, leaving the requestor deadlocked forever
				// to avoid the above,this if is required
				h, err := hs.store.GetByHeight(hs.ctx, req.height)
				req.resp <- &heightResp{h, err} // reqs are always buffered, so this won't block
				continue
			}

			hs.reqs[req.height] = append(hs.reqs[req.height], req.resp)
		case headers := <-hs.provideCh:
			from, to := uint64(headers[0].Height), uint64(headers[len(headers)-1].Height)
			if hs.height != 0 && hs.height+1 != from {
				log.Warnf("BUG: headers given to the heightSub are in the wrong order")
				continue
			}

			// instead of looping over each header in 'headers', we can loop over each request
			// which will drastically decrease idle iterations, as there will be lesser requests than the headers
			for height, reqs := range hs.reqs {
				// then we look if any of the requests match the given range of headers
				if height >= from && height <= to {
					// and if so, calculate its position and fulfill requests
					h := headers[height-from]
					for _, req := range reqs {
						req <- &heightResp{header: h} // reqs are always buffered, so this won't block
					}
					delete(hs.reqs, height)
				}
			}

			hs.height = to
		case <-hs.ctx.Done():
			return
		}
	}
}
