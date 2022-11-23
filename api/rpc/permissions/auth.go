package permissions

import (
	"crypto/rand"
	"io"

	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/gbrlsnchs/jwt/v3"
)

func NewAdminSecret() (*jwt.HMACSHA, error) {
	// TODO @renaynay: implement keystore get token and decode here
	sk, err := io.ReadAll(io.LimitReader(rand.Reader, 32))
	if err != nil {
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
