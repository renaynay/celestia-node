package store

import (
	"context"
	"fmt"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
)

// Init ensures a Store is initialized. If it is not already initialized,
// it initializes the Store by requesting the header with the given hash.
func Init(
	ctx context.Context,
	store header.Store,
	ex header.Exchange,
	hash tmbytes.HexBytes,
	net p2p.Network,
) error {
	_, err := store.Head(ctx)
	switch err {
	default:
		return err
	case header.ErrNoHead:
		initial, err := ex.Get(ctx, hash)
		if err != nil {
			return err
		}
		if initial.RawHeader.ChainID != string(net) {
			return fmt.Errorf("header chainID mismatch, expected %s, got %s", net, initial.RawHeader.ChainID)
		}

		return store.Init(ctx, initial)
	}
}
