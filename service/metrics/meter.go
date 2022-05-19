package metrics

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

// Meter represents
type Meter struct {
	meter metric.Meter

	name string
}

// NewMeter returns a new meter for the given
// name.
func NewMeter(name string) *Meter {
	mtr := global.Meter(name)
	return &Meter{
		meter: mtr,
		name:  name,
	}
}

func (m *Meter)
