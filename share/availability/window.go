package availability

import (
	"errors"
	"time"
)

const (
	RequestWindow = 30 * 24 * time.Hour
	StorageWindow = RequestWindow + time.Hour
)

// TODO @renaynay: describe error
var ErrOutsideSamplingWindow = errors.New("timestamp outside sampling window")

type Window time.Duration

func (w Window) Duration() time.Duration {
	return time.Duration(w)
}

// IsWithinWindow checks whether the given timestamp is within the
// given AvailabilityWindow. If the window is disabled (0), it returns true for
// every timestamp.
func IsWithinWindow(t time.Time, window time.Duration) bool {
	if window == time.Duration(0) { // TODO @renaynay: what to do w this cond  ?
		return true
	}
	return time.Since(t) <= window
}
