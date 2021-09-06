package block

import (
	"context"

	core "github.com/celestiaorg/celestia-core/types"
)

type BlockFetcher interface {
	GetBlock(ctx context.Context, height *int64) (*core.Block, error)
	SubscribeNewBlockEvent(ctx context.Context) (<-chan *core.Block, error)
	UnsubscribeNewBlockEvent(ctx context.Context) error
}

