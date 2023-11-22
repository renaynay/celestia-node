package gateway

import (
	"github.com/celestiaorg/celestia-node/api/gateway"
	"github.com/celestiaorg/celestia-node/das"
	"github.com/celestiaorg/celestia-node/nodebuilder/header"
	"github.com/celestiaorg/celestia-node/nodebuilder/share"
)

// Handler constructs a new RPC Handler from the given services.
func Handler(
	share share.Module,
	header header.Module,
	daser *das.DASer,
	serv *gateway.Server,
) {
	handler := gateway.NewHandler(share, header, daser)
	handler.RegisterEndpoints(serv)
	handler.RegisterMiddleware(serv)
}

func server(cfg *Config) *gateway.Server {
	return gateway.NewServer(cfg.Address, cfg.Port)
}
