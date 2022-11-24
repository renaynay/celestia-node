package client

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/gbrlsnchs/jwt/v3"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-node/api/rpc/permissions"
)

// Tests to impl:
// 1. test client (all perms) against admin endpoint, W + R endpoints
// 2. test client (RONLY) against admin endpoint
// 3. test client (all perms but signed w wrong secret) against admin endpoint, W + R endpoints
// 4. test client no auth header (ensure failure)
// 5. test client malformed auth header (ensure failure)

// TODO @renaynay: bad test
func TestClientPermissions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	filename := "/Users/renenayman/.celestia-light-arabica-2/keys/jwt-Light.jwts"
	file, err := os.Open(filename)
	require.NoError(t, err)
	defer file.Close()

	secret, err := decodeKeyIntoSecret(file)
	require.NoError(t, err)

	token, err := permissions.NewTokenWithPerms(secret, permissions.DefaultPerms)
	require.NoError(t, err)

	cli, err := NewClientWithPerms(ctx, "http://localhost:26658", string(token))
	require.NoError(t, err)

	stats, err := cli.DAS.SamplingStats(ctx)
	require.NoError(t, err)

	t.Log(stats)
}

func TestIt(t *testing.T) {
	cli, err := NewClient(context.Background(), "http://localhost:26658")
	require.NoError(t, err)

	stats, err := cli.DAS.SamplingStats(context.Background())
	require.NoError(t, err)

	t.Log(stats)

}

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
