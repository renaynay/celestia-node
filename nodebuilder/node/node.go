package node

import (
	"fmt"
	"runtime"

	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/gbrlsnchs/jwt/v3"
	logging "github.com/ipfs/go-log/v2"

	"github.com/celestiaorg/celestia-node/api/rpc/permissions"
)

var (
	buildTime       string
	lastCommit      string
	semanticVersion string
)

// Module defines the API related to interacting with the "administrative"
// node.
type Module interface {
	// Type returns the node type.
	Type() Type
	// Version returns information about the current binary build.
	Version() Version

	// LogLevelSet sets the given component log level to the given level.
	LogLevelSet(name, level string) error

	// AuthVerify returns the given token's permissions.
	AuthVerify(token string) ([]auth.Permission, error) //perm:read
	// AuthNew signs and returns a token with the given permissions.
	AuthNew(perms []auth.Permission) ([]byte, error) //perm:admin

}

type API struct {
	Internal struct {
		Type        func() Type                                   `perm:"admin"`
		Version     func() Version                                `perm:"admin"`
		LogLevelSet func(name, level string) error                `perm:"admin"`
		AuthVerify  func(token string) ([]auth.Permission, error) `perm:"admin"`
		AuthNew     func(perms []auth.Permission) ([]byte, error) `perm:"admin"`
	}
}

func (api *API) Type() Type {
	return api.Internal.Type()
}
func (api *API) Version() Version {
	return api.Internal.Version()
}
func (api *API) LogLevelSet(name, level string) error {
	return api.Internal.LogLevelSet(name, level)
}

type admin struct {
	tp Type

	secret *jwt.HMACSHA
}

func newAdmin(tp Type) Module {
	return &admin{
		tp: tp,
	}
}

func (a *admin) Type() Type {
	return a.tp
}

// Version represents all binary build information.
type Version struct {
	SemanticVersion string `json:"semantic_version"`
	LastCommit      string `json:"last_commit"`
	BuildTime       string `json:"build_time"`
	SystemVersion   string `json:"system_version"`
	GoVersion       string `json:"go_version"`
}

func (a *admin) Version() Version {
	return Version{
		SemanticVersion: semanticVersion,
		LastCommit:      lastCommit,
		BuildTime:       buildTime,
		SystemVersion:   fmt.Sprintf("%s/%s", runtime.GOARCH, runtime.GOOS),
		GoVersion:       runtime.Version(),
	}
}

func (a *admin) LogLevelSet(name, level string) error {
	return logging.SetLogLevel(name, level)
}

func (a *admin) AuthVerify(token string) ([]auth.Permission, error) {
	p := &permissions.JWTPayload{}
	_, err := jwt.Verify([]byte(token), a.secret, p)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate token: %w", err)
	}
	return p.Allow, nil
}

func (a *admin) AuthNew(perms []auth.Permission) ([]byte, error) {
	return permissions.NewTokenWithPerms(a.secret, perms)
}
