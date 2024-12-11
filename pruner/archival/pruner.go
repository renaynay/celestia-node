package archival

import (
	"context"

	logging "github.com/ipfs/go-log/v2"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/store"
)

var log = logging.Logger("pruner/archival")

// Pruner is a noop implementation of the pruner.Factory interface
// that allows archival nodes to sync and retain historical data
// that is out of the availability window.
type Pruner struct {
	store *store.Store
}

func NewPruner(store *store.Store) *Pruner {
	return &Pruner{store: store}
}

// Prune prunes the Q4 file related to the block at the given height.
func (p *Pruner) Prune(ctx context.Context, eh *header.ExtendedHeader) error {
}
