package node

import (
	"fmt"
	"runtime"

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
}

type API struct {
	Type        func() Type
	Version     func() Version
	LogLevelSet func(name, level string) error
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
