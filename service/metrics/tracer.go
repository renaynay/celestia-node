package metrics

import (
	logging "github.com/ipfs/go-log/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var TraceLogger = logging.Logger("metrics/tracing")

// NewTracer creates a new tracer for the given package/service name.
func NewTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
