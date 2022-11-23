package rpc

import (
	"github.com/gbrlsnchs/jwt/v3"

	"github.com/celestiaorg/celestia-node/api/rpc"
	"github.com/celestiaorg/celestia-node/nodebuilder/das"
	"github.com/celestiaorg/celestia-node/nodebuilder/fraud"
	"github.com/celestiaorg/celestia-node/nodebuilder/header"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/share"
	"github.com/celestiaorg/celestia-node/nodebuilder/state"
)

// RegisterEndpoints registers the given services on the rpc.
func RegisterEndpoints(
	stateMod state.Module,
	shareMod share.Module,
	fraudMod fraud.Module,
	headerMod header.Module,
	daserMod das.Module,
	nodeMod node.Module,
	serv *rpc.Server,
) {
	serv.RegisterAuthedService("state", stateMod, &state.API{})
	serv.RegisterAuthedService("share", shareMod, &share.API{})
	serv.RegisterAuthedService("fraud", fraudMod, &fraud.API{})
	serv.RegisterAuthedService("header", headerMod, &header.API{})
	serv.RegisterAuthedService("das", daserMod, &das.API{})
	serv.RegisterAuthedService("node", nodeMod, &node.API{})
}

func Server(cfg *Config, auth *jwt.HMACSHA) *rpc.Server {
	return rpc.NewServer(cfg.Address, cfg.Port, auth)
}
