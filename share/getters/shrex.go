package getters

import (
	"context"
	"errors"

	"github.com/celestiaorg/celestia-node/share"
	"github.com/celestiaorg/celestia-node/share/p2p"
	"github.com/celestiaorg/celestia-node/share/p2p/peers"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexeds"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexnd"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexsub"

	"github.com/celestiaorg/nmt/namespace"
	"github.com/celestiaorg/rsmt2d"
)

var _ share.Getter = (*ShrexGetter)(nil)

// ShrexGetter is a share.Getter that uses the shrex/eds and shrex/nd protocol to retrieve shares.
type ShrexGetter struct {
	edsClient *shrexeds.Client
	ndClient  *shrexnd.Client
	shrexSub  *shrexsub.PubSub

	peers *peers.Manager
}

func NewShrexGetter(
	edsClient *shrexeds.Client,
	ndClient *shrexnd.Client,
	shrexSub *shrexsub.PubSub,
	peerManager *peers.Manager,
) *ShrexGetter {
	return &ShrexGetter{
		edsClient: edsClient,
		ndClient:  ndClient,
		shrexSub:  shrexSub,
		peers:     peerManager,
	}
}

func (sg *ShrexGetter) Start(ctx context.Context) error {
	sg.peers.Start()
	// TODO: Do we need to subscribe?
	return sg.shrexSub.AddValidator(sg.peers.Validate)
}

func (sg *ShrexGetter) Stop(ctx context.Context) error {
	return sg.peers.Stop(ctx)
}

func (sg *ShrexGetter) GetShare(ctx context.Context, root *share.Root, row, col int) (share.Share, error) {
	return nil, errors.New("shrex-getter: GetShare is not supported")
}

func (sg *ShrexGetter) GetEDS(ctx context.Context, root *share.Root) (*rsmt2d.ExtendedDataSquare, error) {
	for {
		to, setStatus, err := sg.peers.GetPeer(ctx, root.Hash())
		if err != nil {
			return nil, err
		}

		eds, err := sg.edsClient.RequestEDS(ctx, root.Hash(), to)
		switch err {
		case nil:
			setStatus(peers.ResultSuccess)
			return eds, nil
		case p2p.ErrInvalidResponse:
			setStatus(peers.ResultPeerMMisbehaved)
		default:
			setStatus(peers.ResultFail)
		}
	}
}

func (sg *ShrexGetter) GetSharesByNamespace(
	ctx context.Context,
	root *share.Root,
	id namespace.ID,
) (share.NamespacedShares, error) {
	for {
		to, setStatus, err := sg.peers.GetPeer(ctx, root.Hash())
		if err != nil {
			return nil, err
		}

		eds, err := sg.ndClient.RequestND(ctx, root, id, to)
		switch err {
		case nil:
			setStatus(peers.ResultSuccess)
			return eds, nil
		case p2p.ErrInvalidResponse:
			setStatus(peers.ResultPeerMMisbehaved)
		default:
			setStatus(peers.ResultFail)
		}
	}
}
