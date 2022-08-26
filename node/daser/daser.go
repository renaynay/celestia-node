package daser

import (
	"github.com/ipfs/go-datastore"

	"github.com/celestiaorg/celestia-node/das"
	"github.com/celestiaorg/celestia-node/fraud"
	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/service/share"
)

// DASer constructs a new Data Availability Sampler.
func DASer(
	avail share.Availability,
	sub header.Subscriber,
	hstore header.Store,
	ds datastore.Batching,
	fservice fraud.Service,
) *das.DASer {
	return das.NewDASer(avail, sub, hstore, ds, fservice)
}
