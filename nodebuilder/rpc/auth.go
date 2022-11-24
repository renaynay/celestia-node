package rpc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gbrlsnchs/jwt/v3"
	"io"
	"os"
	"strings"

	"github.com/celestiaorg/celestia-node/api/rpc/permissions"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
)

func getSecret(tp node.Type, path string) (*jwt.HMACSHA, error) { // TODO @renaynay eventually this should be keystore
	ksPath := path + "/keys"
	// check in keystore
	file, err := os.Open(ksPath)
	if os.IsExist(err) {
		defer file.Close()
		return decodeKeyIntoSecret(file)
	}
	return generateNewWithSignedToken(tp, path)
}

// decodeKeyIntoSecret // TODO @renaynay
func decodeKeyIntoSecret(input io.Reader) (*jwt.HMACSHA, error) {
	encoded, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}

	decoded, err := hex.DecodeString(strings.TrimSpace(string(encoded)))
	if err != nil {
		return nil, err
	}

	var keyInfo permissions.KeyInfo
	if err := json.Unmarshal(decoded, &keyInfo); err != nil {
		return nil, err
	}

	return jwt.NewHS256(keyInfo.PrivKey), nil
}

func generateNewWithSignedToken(tp node.Type, path string) (*jwt.HMACSHA, error) {
	// generate new JWT secret and save
	secret, err := permissions.NewAdminSecret(tp.String(), path)
	if err != nil {
		return nil, err
	}

	// generate admin token from secret and save to same path
	p := permissions.JWTPayload{
		Allow: permissions.AllPerms,
	}
	token, err := jwt.Sign(&p, secret)
	if err != nil {
		return nil, err
	}
	filenameToken := fmt.Sprintf("%s/jwt-%s.token", path, tp.String())
	err = os.WriteFile(filenameToken, token, 0600)
	if err != nil {
		return nil, err
	}

	return secret, nil
}
