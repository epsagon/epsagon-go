package internal

import (
	"context"

	"github.com/epsagon/epsagon-go/tracer"
)

// Extracts the tracer from given contexts (using first context),
// returns Global tracer if no context is given and GlobalTracer is valid (= non nil, not stopped)
func ExtractTracer(ctx []context.Context) tracer.Tracer {
	if len(ctx) == 0 {
		if tracer.GlobalTracer == nil || tracer.GlobalTracer.Stopped() {
			return nil
		}
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
	if tracerValue == nil || tracerValue.Stopped() {
		return nil
	}
	return tracerValue
}
