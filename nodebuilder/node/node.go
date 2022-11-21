package node

import (
	"context"
	"fmt"
	"runtime"

	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/golang-jwt/jwt/v4"
	logging "github.com/ipfs/go-log/v2"
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

	AuthVerify(ctx context.Context, token string) ([]auth.Permission, error) //perm:read
	AuthNew(ctx context.Context, perms []auth.Permission) ([]byte, error)    //perm:admin

}

type API struct {
	Internal struct {
		Type        func() Type                                              `perm:"admin"`
		Version     func() Version                                           `perm:"admin"`
		LogLevelSet func(name, level string) error                           `perm:"admin"`
		AuthVerify  func(context.Context, string) ([]auth.Permission, error) `perm:"admin"`
		AuthNew     func(ctx context.Context, perms []auth.Permission) ([]byte, error)
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

func (a *admin) AuthVerify(ctx context.Context, token string) ([]auth.Permission, error) {
	//TODO implement me
	panic("implement me")
}

func (a *admin) AuthNew(ctx context.Context, perms []auth.Permission) ([]byte, error) {
	jwt.New()

	//TODO implement me
	panic("implement me")
}
