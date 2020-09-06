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
	return ctx[0].Value("tracer").(tracer.Tracer)
}
