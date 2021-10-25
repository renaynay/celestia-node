package header

import "context"

// Subscriber encompasses the behaviour necessary for a node to
// subscribe/unsubscribe from new ExtendedHeader events from the
// network.
type Subscriber interface {
	Subscribe(ctx context.Context) (<-chan *ExtendedHeader, error)
	Unsubscribe() error
}

// Exchange encompasses the behaviour necessary for a node to
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
