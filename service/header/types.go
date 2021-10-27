package header

import (
	mrand "math/rand"
	"time"

	tmrand "github.com/celestiaorg/celestia-core/libs/rand"
	"github.com/celestiaorg/celestia-core/pkg/da"
	"github.com/celestiaorg/celestia-core/proto/tendermint/version"
	core "github.com/celestiaorg/celestia-core/types"
)

// ExtendedHeader represents a wrapped "raw" header that includes
// information necessary for Celestia Nodes to be notified of new
// block headers and perform Data Availability Sampling.
type ExtendedHeader struct {
	*RawHeader
	DAH *da.DataAvailabilityHeader
}

// RawHeader is an alias to core.Header. It is
// "raw" because it is not yet wrapped to include
// the DataAvailabilityHeader.
type RawHeader = core.Header

// ExtendedHeaderRequest represents a request for one or several
// ExtendedHeaders.
type ExtendedHeaderRequest struct {
	Origin  int64 // TODO @renaynay: should we do this by hash or height?
	Amount  int64 // TODO @renaynay: should this just be an uint?
	Reverse bool
}

func (e *ExtendedHeader) Marshal() []byte { // TODO @renaynay: delete these once rebased on hlib's proto PR
	return make([]byte, 32)
}

func (e *ExtendedHeader) Unmarshal([]byte) { // TODO @renaynay: delete these once rebased on hlib's proto PR
	bid := core.BlockID{
		Hash: make([]byte, 32),
		PartSetHeader: core.PartSetHeader{
			Total: 123,
			Hash:  make([]byte, 32),
		},
	}
	mrand.Read(bid.Hash)               //nolint:gosec
	mrand.Read(bid.PartSetHeader.Hash) //nolint:gosec

	e.Version = version.Consensus{Block: 11, App: 1}
	e.ChainID = "test"
	e.Height = mrand.Int63() //nolint:gosec
	e.Time = time.Now()
	e.LastBlockID = bid
	e.LastCommitHash = tmrand.Bytes(32)
	e.DataHash = tmrand.Bytes(32)
	e.ValidatorsHash = tmrand.Bytes(32)
	e.NextValidatorsHash = tmrand.Bytes(32)
	e.ConsensusHash = tmrand.Bytes(32)
	e.AppHash = tmrand.Bytes(32)
	e.LastResultsHash = tmrand.Bytes(32)
	e.EvidenceHash = tmrand.Bytes(32)
	e.ProposerAddress = tmrand.Bytes(20)
}
