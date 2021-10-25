package header

import "context"

// Exchange encompasses the behaviour necessary for a Node to
// retrieve and provide ExtendedHeaders to the Celestia network.
type Exchange interface {
	RetrieveHeaders(ctx context.Context, request *ExtendedHeaderRequest) ([]*ExtendedHeader, error)
	ProvideHeaders(ctx context.Context, header []*ExtendedHeader) error
}

// Store encompasses the behaviour necessary to store and retrieve ExtendedHeaders
// from a node's local storage.
type Store interface {
	GetHeaders(ctx context.Context, request *ExtendedHeaderRequest) ([]*ExtendedHeader, error)
	StoreHeaders(ctx context.Context, headers []*ExtendedHeader) error
}
