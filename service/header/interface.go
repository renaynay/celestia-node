package header

import "context"

// Exchange encompasses the behaviour necessary for a Node to
// retrieve and provide ExtendedHeaders to the Celestia network.
type Exchange interface {
	Retrieve(ctx context.Context, request *ExtendedHeaderRequest) ([]*ExtendedHeader, error)
	Provide(ctx context.Context, header *ExtendedHeader) error
}
