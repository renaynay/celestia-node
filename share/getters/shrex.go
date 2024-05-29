package getters

import (
	"context"
	"errors"
	"fmt"
	"time"

	libpeer "github.com/libp2p/go-libp2p/core/peer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/celestiaorg/rsmt2d"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/libs/utils"
	"github.com/celestiaorg/celestia-node/pruner"
	"github.com/celestiaorg/celestia-node/share"
	"github.com/celestiaorg/celestia-node/share/ipld"
	"github.com/celestiaorg/celestia-node/share/p2p"
	"github.com/celestiaorg/celestia-node/share/p2p/peers"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexeds"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexnd"
)

var _ share.Getter = (*ShrexGetter)(nil)

const (
	// defaultMinRequestTimeout value is set according to observed time taken by healthy peer to
	// serve getEDS request for block size 256
	defaultMinRequestTimeout = time.Minute // should be >= shrexeds server write timeout
	defaultMinAttemptsCount  = 3
)

var meter = otel.Meter("shrex/getter")

type metrics struct {
	edsAttempts metric.Int64Histogram
	ndAttempts  metric.Int64Histogram
}

func (m *metrics) recordEDSAttempt(ctx context.Context, attemptCount int, success bool) {
	if m == nil {
		return
	}
	ctx = utils.ResetContextOnError(ctx)
	m.edsAttempts.Record(ctx, int64(attemptCount),
		metric.WithAttributes(
			attribute.Bool("success", success)))
}

func (m *metrics) recordNDAttempt(ctx context.Context, attemptCount int, success bool) {
	if m == nil {
		return
	}
	ctx = utils.ResetContextOnError(ctx)
	m.ndAttempts.Record(ctx, int64(attemptCount),
		metric.WithAttributes(
			attribute.Bool("success", success)))
}

func (sg *ShrexGetter) WithMetrics() error {
	edsAttemptHistogram, err := meter.Int64Histogram(
		"getters_shrex_eds_attempts_per_request",
		metric.WithDescription("Number of attempts per shrex/eds request"),
	)
	if err != nil {
		return err
	}

	ndAttemptHistogram, err := meter.Int64Histogram(
		"getters_shrex_nd_attempts_per_request",
		metric.WithDescription("Number of attempts per shrex/nd request"),
	)
	if err != nil {
		return err
	}

	sg.metrics = &metrics{
		edsAttempts: edsAttemptHistogram,
		ndAttempts:  ndAttemptHistogram,
	}
	return nil
}

// ShrexGetter is a share.Getter that uses the shrex/eds and shrex/nd protocol to retrieve shares.
type ShrexGetter struct {
	edsClient *shrexeds.Client
	ndClient  *shrexnd.Client

	fullPeerManager     *peers.Manager
	archivalPeerManager *peers.Manager

	// minRequestTimeout limits minimal timeout given to single peer by getter for serving the request.
	minRequestTimeout time.Duration
	// minAttemptsCount will be used to split request timeout into multiple attempts. It will allow to
	// attempt multiple peers in scope of one request before context timeout is reached
	minAttemptsCount int

	availabilityWindow pruner.AvailabilityWindow

	metrics *metrics
}

func NewShrexGetter(
	edsClient *shrexeds.Client,
	ndClient *shrexnd.Client,
	fullPeerManager *peers.Manager,
	archivalManager *peers.Manager,
	availWindow pruner.AvailabilityWindow,
) *ShrexGetter {
	s := &ShrexGetter{
		edsClient:           edsClient,
		ndClient:            ndClient,
		fullPeerManager:     fullPeerManager,
		archivalPeerManager: archivalManager,
		minRequestTimeout:   defaultMinRequestTimeout,
		minAttemptsCount:    defaultMinAttemptsCount,
		availabilityWindow:  availWindow,
	}

	return s
}

func (sg *ShrexGetter) Start(ctx context.Context) error {
	err := sg.fullPeerManager.Start(ctx)
	if err != nil {
		return err
	}
	return sg.archivalPeerManager.Start(ctx)
}

func (sg *ShrexGetter) Stop(ctx context.Context) error {
	err := sg.fullPeerManager.Stop(ctx)
	if err != nil {
		return err
	}
	return sg.archivalPeerManager.Stop(ctx)
}

func (sg *ShrexGetter) GetShare(context.Context, *header.ExtendedHeader, int, int) (share.Share, error) {
	return nil, fmt.Errorf("getter/shrex: GetShare %w", errOperationNotSupported)
}

func (sg *ShrexGetter) GetEDS(ctx context.Context, header *header.ExtendedHeader) (*rsmt2d.ExtendedDataSquare, error) {
	var err error
	ctx, span := tracer.Start(ctx, "shrex/get-eds")
	defer func() {
		utils.SetStatusAndEnd(span, err)
	}()

	// short circuit if the data root is empty
	if header.DAH.Equals(share.EmptyRoot()) {
		return share.EmptyExtendedDataSquare(), nil
	}

	var attempt int
	for {
		if ctx.Err() != nil {
			sg.metrics.recordEDSAttempt(ctx, attempt, false)
			return nil, errors.Join(err, ctx.Err())
		}
		attempt++
		start := time.Now()

		peer, setStatus, getErr := sg.getPeer(ctx, header)
		if getErr != nil {
			log.Debugw("eds: couldn't find peer",
				"hash", header.DAH.String(),
				"err", getErr,
				"finished (s)", time.Since(start))
			sg.metrics.recordEDSAttempt(ctx, attempt, false)
			return nil, errors.Join(err, getErr)
		}

		reqStart := time.Now()
		reqCtx, cancel := ctxWithSplitTimeout(ctx, sg.minAttemptsCount-attempt+1, sg.minRequestTimeout)
		eds, getErr := sg.edsClient.RequestEDS(reqCtx, header.DAH.Hash(), peer)
		cancel()
		switch {
		case getErr == nil:
			setStatus(peers.ResultNoop)
			sg.metrics.recordEDSAttempt(ctx, attempt, true)
			return eds, nil
		case errors.Is(getErr, context.DeadlineExceeded),
			errors.Is(getErr, context.Canceled):
			setStatus(peers.ResultCooldownPeer)
		case errors.Is(getErr, p2p.ErrNotFound):
			getErr = share.ErrNotFound
			setStatus(peers.ResultCooldownPeer)
		case errors.Is(getErr, p2p.ErrInvalidResponse):
			setStatus(peers.ResultBlacklistPeer)
		default:
			setStatus(peers.ResultCooldownPeer)
		}

		if !ErrorContains(err, getErr) {
			err = errors.Join(err, getErr)
		}
		log.Debugw("eds: request failed",
			"hash", header.DAH.String(),
			"peer", peer.String(),
			"attempt", attempt,
			"err", getErr,
			"finished (s)", time.Since(reqStart))
	}
}

func (sg *ShrexGetter) GetSharesByNamespace(
	ctx context.Context,
	header *header.ExtendedHeader,
	namespace share.Namespace,
) (share.NamespacedShares, error) {
	if err := namespace.ValidateForData(); err != nil {
		return nil, err
	}
	var (
		attempt int
		err     error
	)
	ctx, span := tracer.Start(ctx, "shrex/get-shares-by-namespace", trace.WithAttributes(
		attribute.String("namespace", namespace.String()),
	))
	defer func() {
		utils.SetStatusAndEnd(span, err)
	}()

	// verify that the namespace could exist inside the roots before starting network requests
	dah := header.DAH
	roots := ipld.FilterRootByNamespace(dah, namespace)
	if len(roots) == 0 {
		return []share.NamespacedRow{}, nil
	}

	for {
		if ctx.Err() != nil {
			sg.metrics.recordNDAttempt(ctx, attempt, false)
			return nil, errors.Join(err, ctx.Err())
		}
		attempt++
		start := time.Now()

		peer, setStatus, getErr := sg.getPeer(ctx, header)
		if getErr != nil {
			log.Debugw("nd: couldn't find peer",
				"hash", dah.String(),
				"namespace", namespace.String(),
				"err", getErr,
				"finished (s)", time.Since(start))
			sg.metrics.recordNDAttempt(ctx, attempt, false)
			return nil, errors.Join(err, getErr)
		}

		reqStart := time.Now()
		reqCtx, cancel := ctxWithSplitTimeout(ctx, sg.minAttemptsCount-attempt+1, sg.minRequestTimeout)
		nd, getErr := sg.ndClient.RequestND(reqCtx, dah, namespace, peer)
		cancel()
		switch {
		case getErr == nil:
			// both inclusion and non-inclusion cases needs verification
			if verErr := nd.Verify(dah, namespace); verErr != nil {
				getErr = verErr
				setStatus(peers.ResultBlacklistPeer)
				break
			}
			setStatus(peers.ResultNoop)
			sg.metrics.recordNDAttempt(ctx, attempt, true)
			return nd, nil
		case errors.Is(getErr, context.DeadlineExceeded),
			errors.Is(getErr, context.Canceled):
			setStatus(peers.ResultCooldownPeer)
		case errors.Is(getErr, p2p.ErrNotFound):
			getErr = share.ErrNotFound
			setStatus(peers.ResultCooldownPeer)
		case errors.Is(getErr, p2p.ErrInvalidResponse):
			setStatus(peers.ResultBlacklistPeer)
		default:
			setStatus(peers.ResultCooldownPeer)
		}

		if !ErrorContains(err, getErr) {
			err = errors.Join(err, getErr)
		}
		log.Debugw("nd: request failed",
			"hash", dah.String(),
			"namespace", namespace.String(),
			"peer", peer.String(),
			"attempt", attempt,
			"err", getErr,
			"finished (s)", time.Since(reqStart))
	}
}

func (sg *ShrexGetter) getPeer(
	ctx context.Context,
	header *header.ExtendedHeader,
) (libpeer.ID, peers.DoneFunc, error) {
	if !pruner.IsWithinAvailabilityWindow(header.Time(), sg.availabilityWindow) {
		log.Infow("REQUESTING HISTORIC FROM ARCHIVAL PEER")
		p, df, err := sg.archivalPeerManager.Peer(ctx, header.DAH.Hash(), header.Height())
		return p, df, err
	}
	log.Infow("REQUESTING RECENT FROM FULL PEER")
	return sg.fullPeerManager.Peer(ctx, header.DAH.Hash(), header.Height())
}
