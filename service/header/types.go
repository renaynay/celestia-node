package header

import (
	"github.com/celestiaorg/celestia-core/pkg/da"
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
