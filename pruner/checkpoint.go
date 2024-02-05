package pruner

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/ipfs/go-datastore"

	"github.com/celestiaorg/celestia-node/header"
)

var (
	storePrefix   = datastore.NewKey("pruner")
	checkpointKey = datastore.NewKey("checkpoint")
)

// checkpoint contains information related to the state of the
// pruner service that is periodically persisted to disk.
type checkpoint struct {
	lastPrunedHeader atomic.Pointer[header.ExtendedHeader]

	LastPrunedHeight uint64            `json:"last_pruned_height"`
	FailedHeaders    map[uint64]string `json:"failed,omitempty"`
}

// initializeCheckpoint initializes the checkpoint, storing the earliest header in the chain.
func (s *Service) initializeCheckpoint(ctx context.Context) error {
	firstHeader, err := s.getter.GetByHeight(ctx, 1)
	if err != nil {
		return fmt.Errorf("failed to initialize checkpoint: %w", err)
	}

	return s.updateCheckpoint(ctx, firstHeader, nil)
}

// loadCheckpoint loads the last checkpoint from disk, initializing it if it does not already exist.
func (s *Service) loadCheckpoint(ctx context.Context) error {
	bin, err := s.ds.Get(ctx, checkpointKey)
	if err != nil {
		if err == datastore.ErrNotFound {
			return s.initializeCheckpoint(ctx)
		}
		return fmt.Errorf("failed to load checkpoint: %w", err)
	}

	var cp *checkpoint
	err = json.Unmarshal(bin, &cp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}
	s.checkpoint = cp

	// load last pruned header based off height
	lastPruned, err := s.getter.GetByHeight(ctx, cp.LastPrunedHeight)
	if err != nil {
		return fmt.Errorf("failed to load last pruned header at height %d: %w", cp.LastPrunedHeight, err)
	}

	s.checkpoint.lastPrunedHeader.Store(lastPruned)
	return nil
}

// updateCheckpoint updates the checkpoint with the last pruned header height
// and persists it to disk.
func (s *Service) updateCheckpoint(
	ctx context.Context,
	lastPruned *header.ExtendedHeader,
	failedHeights map[uint64]error,
) error {
	for height, failErr := range failedHeights {
		// if the height already exists, just update the error
		s.checkpoint.FailedHeaders[height] = failErr.Error()
	}

	s.checkpoint.lastPrunedHeader.Store(lastPruned)
	s.checkpoint.LastPrunedHeight = lastPruned.Height()

	bin, err := json.Marshal(s.checkpoint)
	if err != nil {
		return err
	}

	return s.ds.Put(ctx, checkpointKey, bin)
}

func (s *Service) lastPruned() *header.ExtendedHeader {
	return s.checkpoint.lastPrunedHeader.Load()
}
