package gateway

import (
	logging "github.com/ipfs/go-log/v2"

	"github.com/celestiaorg/celestia-node/das"
	"github.com/celestiaorg/celestia-node/nodebuilder/header"
	"github.com/celestiaorg/celestia-node/nodebuilder/share"
)

var log = logging.Logger("gateway")

type Handler struct {
	share  share.Module
	header header.Module
	das    *das.DASer
}

func NewHandler(
	share share.Module,
	header header.Module,
	das *das.DASer,
) *Handler {
	return &Handler{
		share:  share,
		header: header,
		das:    das,
	}
}
