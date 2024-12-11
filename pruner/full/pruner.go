package full

import (
	"context"

	logging "github.com/ipfs/go-log/v2"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/store"
)

var log = logging.Logger("pruner/full")

type Pruner struct {
	store *store.Store
}

func NewPruner(store *store.Store) *Pruner {
	return &Pruner{
		store: store,
	}
}

func (p *Pruner) Prune(ctx context.Context, eh *header.ExtendedHeader) error {
}
