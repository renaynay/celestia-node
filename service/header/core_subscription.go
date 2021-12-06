package header

import (
	"context"
	"errors"
	"fmt"

	"github.com/tendermint/tendermint/types"
)

type coreSubscription struct {
	ex *CoreExchange

	sub <-chan *types.Block
}

func newCoreSubscription(ex Exchange) (*coreSubscription, error) {
	coreEx, ok := ex.(*CoreExchange)
	if !ok {
		// TODO @renaynay: better error?
		return nil, errors.New("header using core backend must use header.coreExchange implementation")
	}
	sub, err := coreEx.fetcher.SubscribeNewBlockEvent(context.Background())
	if err != nil {
		return nil, err
	}

	return &coreSubscription{
		ex:  coreEx,
		sub: sub,
	}, nil
}

func (cs *coreSubscription) NextHeader(ctx context.Context) (*ExtendedHeader, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	case newBlock, ok := <-cs.sub:
		if !ok {
			return nil, fmt.Errorf("subscription closed")
		}
		return cs.ex.generateExtendedHeaderFromBlock(newBlock)
	}
}

func (cs *coreSubscription) Cancel() error {
	return cs.ex.fetcher.UnsubscribeNewBlockEvent(context.Background())
}
