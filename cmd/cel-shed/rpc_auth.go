package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gbrlsnchs/jwt/v3"
	"github.com/spf13/cobra"

	"github.com/celestiaorg/celestia-node/api/rpc/permissions"
)

func init() {
	authCmd.AddCommand(authNewSecretCmd, authNewTokenCmd)
}

var authCmd = &cobra.Command{
	Use:   "auth [subcommand]",
	Short: "Collection of auth-related utilities",
}

var authNewSecretCmd = &cobra.Command{
	// TODO @renaynay: eventually this logic will be moved to node admin api and this cli tool will call it
	Use:   "new-secret [name] [path]",
	Short: "Generate new JWT secret with the given name and save to given path",
	RunE: func(cmd *cobra.Command, args []string) error {
		return NewJWTSecret(args)
	},
}

var authNewTokenCmd = &cobra.Command{
	// TODO @renaynay: eventually this logic will be moved to node admin api and this cli tool will call it
	Use:   "new-token [name] [path]",
	Short: "Generate new signed JWT token with the given name using the secret at the given path",
	RunE: func(cmd *cobra.Command, args []string) error {
		return NewJWTSecret(args)
	},
}

type KeyInfo struct {
	PrivKey []byte
}

func NewJWTSecret(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("must specify name and path")
	}
	name := args[0]
	// sanity-check path
	path, err := filepath.Abs(args[1])
	if err != nil {
		return err
	}

	// generate new JWT secret and save
	sk, err := io.ReadAll(io.LimitReader(rand.Reader, 32))
	if err != nil {
		return err
	}
	ki := &KeyInfo{
		PrivKey: sk,
	}

	filename := fmt.Sprintf("%s/jwt-%s.jwts", path, name)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Error closing file: %w", err)
		}
	}()
	bytes, err := json.Marshal(ki)
	if err != nil {
		return err
	}
	encoded := hex.EncodeToString(bytes)
	if _, err := file.Write([]byte(encoded)); err != nil {
		return err
	}

	// generate token from secret and save
	p := permissions.JWTPayload{
		Allow: permissions.AllPerms,
	}
	token, err := jwt.Sign(&p, jwt.NewHS256(sk))
	if err != nil {
		return err
	}
	filenameToken := fmt.Sprintf("%s/jwt-%s.token", path, name)
	return os.WriteFile(filenameToken, token, 0600)
}

// TODO @renaynay: eventually add ability to generate new token with same key
func SignToken(args []string) error {
	return nil
}
