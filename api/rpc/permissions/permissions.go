package permissions

import "github.com/filecoin-project/go-jsonrpc/auth"

var (
	AllPerms     = []auth.Permission{"read", "write", "admin"}
	DefaultPerms = []auth.Permission{"read"}
)
