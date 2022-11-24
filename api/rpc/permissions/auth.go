package permissions

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/gbrlsnchs/jwt/v3"
)

type KeyInfo struct { // TODO @renaynay dedup
	PrivKey []byte
}

func NewAdminSecret(name, path string) (*jwt.HMACSHA, error) {
	// TODO @renaynay: implement keystore get token and decode here
	sk, err := io.ReadAll(io.LimitReader(rand.Reader, 32))
	if err != nil {
		return nil, err
	}

	ki := &KeyInfo{
		PrivKey: sk,
	}

	// TODO @renaynay: It doesn't make sense to generate the jwt secret without saving it anywhere but idk
	// what this API should look like
	filename := fmt.Sprintf("%s/jwt-%s.jwts", path, name)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Error closing file: %w", err)
		}
	}()

	bytes, err := json.Marshal(ki)
	if err != nil {
		return nil, err
	}

	encoded := hex.EncodeToString(bytes)
	if _, err := file.Write([]byte(encoded)); err != nil {
		return nil, err
	}

	return jwt.NewHS256(sk), nil
}

func NewTokenWithPerms(secret *jwt.HMACSHA, perms []auth.Permission) ([]byte, error) {
	p := JWTPayload{
		Allow: perms,
	}
	return jwt.Sign(&p, secret)
}
