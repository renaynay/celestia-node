package header

import (
	"context"

	"github.com/celestiaorg/celestia-node/header"
	libhead "github.com/celestiaorg/celestia-node/libs/header"
	"github.com/celestiaorg/celestia-node/libs/header/p2p"
	"github.com/celestiaorg/celestia-node/libs/header/sync"
)

// Service represents the header Service that can be started / stopped on a node.
// Service's main function is to manage its sub-services. Service can contain several
// sub-services, such as Exchange, ExchangeServer, Syncer, and so forth.
type Service struct {
	ex libhead.Exchange[*header.ExtendedHeader]

	syncer    *sync.Syncer[*header.ExtendedHeader]
	sub       libhead.Subscriber[*header.ExtendedHeader]
	p2pServer *p2p.ExchangeServer[*header.ExtendedHeader]
	store     libhead.Store[*header.ExtendedHeader]
}

// newHeaderService creates a new instance of header Service.
func newHeaderService(
	syncer *sync.Syncer[*header.ExtendedHeader],
	sub libhead.Subscriber[*header.ExtendedHeader],
	p2pServer *p2p.ExchangeServer[*header.ExtendedHeader],
	ex libhead.Exchange[*header.ExtendedHeader],
	store libhead.Store[*header.ExtendedHeader]) Module {
	return &Service{
		syncer:    syncer,
		sub:       sub,
		p2pServer: p2pServer,
		ex:        ex,
		store:     store,
	}
}

func (s *Service) Get(ctx context.Context, hash libhead.Hash) (*header.ExtendedHeader, error) {
	return s.store.Get(ctx, hash)
}

func (s *Service) GetVerifiedRangeByHeight(
	ctx context.Context,
	from *header.ExtendedHeader,
	to uint64,
) ([]*header.ExtendedHeader, error) {
	return s.store.GetVerifiedRange(ctx, from, to)
}

func (s *Service) GetByHeight(ctx context.Context, height uint64) (*header.ExtendedHeader, error) {
	return s.store.GetByHeight(ctx, height)
}

func (s *Service) Head(ctx context.Context) (*header.ExtendedHeader, error) {
	return s.store.Head(ctx)
}

func (s *Service) IsSyncing(context.Context) bool {
	return !s.syncer.State().Finished()
}

func (s *Service) WaitSync(ctx context.Context) error {
	return s.syncer.WaitSync(ctx)
}

func (s *Service) SyncHead(ctx context.Context) (*header.ExtendedHeader, error) {
	return s.syncer.Head(ctx)
}

func (s *Service) Subscribe(context.Context) (libhead.Subscription[*header.ExtendedHeader], error) {
	return s.sub.Subscribe()
}
