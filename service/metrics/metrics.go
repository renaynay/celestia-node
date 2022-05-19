package metrics

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// NewTracer creates a new tracer for the given package/service name.
func NewTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

func insertUser(ctx context.Context, user *User) error {
	ctx, span := tracer.Start(ctx, "insert-user")
	defer span.End()
}

