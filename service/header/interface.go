package header

import "context"

// Subscriber encompasses the behavior necessary to
// subscribe/unsubscribe from new ExtendedHeader events from the
// network.
type Subscriber interface {
	Subscribe() (Subscription, error)
}

type Subscription interface {
	NextHeader(ctx context.Context) (*ExtendedHeader, error)
	Cancel()
}

// Exchange encompasses the behavior necessary for a node to
// request ExtendedHeaders and respond to ExtendedHeader requests
// from the network.
type Exchange interface {
	RequestHeaders(ctx context.Context, request *ExtendedHeaderRequest) ([]*ExtendedHeader, error)
	RespondToHeadersRequest(ctx context.Context, request *ExtendedHeaderRequest) error
}

// Store encompasses the behavior necessary to store and retrieve ExtendedHeaders
// from a node's local storage.
type Store interface {
	GetHeaders(ctx context.Context, request *ExtendedHeaderRequest) ([]*ExtendedHeader, error)
	StoreHeaders(ctx context.Context, headers []*ExtendedHeader) error
}
