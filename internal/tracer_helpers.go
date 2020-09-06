package internal

import (
	"context"
	"github.com/epsagon/epsagon-go/tracer"
)

// Extracts the tracer from given contexts (using first context),
// returns Global tracer if no context is given
func ExtractTracer(ctx []context.Context) tracer.Tracer {
	if len(ctx) == 0 {
		return tracer.GlobalTracer
	}
	rawValue := ctx[0].Value("tracer")
	if rawValue == nil {
		panic("Invalid context, see Epsagon Concurrent Generic GO function example")
	}
	tracerValue, ok := rawValue.(tracer.Tracer)
	if !ok {
		panic("Invalid context value, see Epsagon Concurrent Generic GO function example")
	}
	return tracerValue
}
