package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/gbrlsnchs/jwt/v3"
	"os"
	"strings"

	"github.com/celestiaorg/celestia-node/api/rpc"
	"github.com/celestiaorg/celestia-node/nodebuilder/das"
	"github.com/celestiaorg/celestia-node/nodebuilder/fraud"
	"github.com/celestiaorg/celestia-node/nodebuilder/header"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/share"
	"github.com/celestiaorg/celestia-node/nodebuilder/state"
)

const authKey = "rpc_auth"

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

func Server(cfg *Config) *rpc.Server {
	return rpc.NewServer(cfg.Address, cfg.Port)
}

func auth(keystorePath string) (*jwt.HMACSHA, error) {
	os.Open(fmt.Sprintf("%s/%s", keystorePath, authKey))

	decoded, err := hex.DecodeString(strings.TrimSpace(string(encoded)))
	if err != nil {
		return err
	}

	jwt.NewHS256()
}
