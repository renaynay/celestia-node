package full

import (
	"time"

	"github.com/celestiaorg/celestia-node/pruner"
)

// Window is the availability window for light nodes in the Celestia
// network (30 days).
const Window = pruner.AvailabilityWindow(time.Minute * 30)