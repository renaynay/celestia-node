package block

import (
	"context"

	ipld "github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log/v2"
)

// TODO @renaynay: document
type Service struct {
	store       ipld.DAGService
}

var log = logging.Logger("block-service")

// NewBlockService creates a new instance of block Service.
func NewBlockService(store ipld.DAGService) *Service {
	return &Service{
		store:       store,
	}
}

// Start starts the block Service.
func (s *Service) Start(ctx context.Context) error {
	log.Info("starting block service")
	return nil
}

// Stop stops the block Service.
func (s *Service) Stop(ctx context.Context) error {
	log.Info("stopping block service")
	return nil
}
