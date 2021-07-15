package epsagon

import (
	"context"

	"github.com/epsagon/epsagon-go/tracer"
)

type tracerKey string

const tracerKeyValue tracerKey = "tracer"

// ContextWithTracer creates a context with given tracer
func ContextWithTracer(t tracer.Tracer, ctx ...context.Context) context.Context {
	if len(ctx) == 1 {
		return context.WithValue(ctx[0], tracerKeyValue, t)
	}
	return context.WithValue(context.Background(), tracerKeyValue, t)
}

// ExtractTracer Extracts the tracer from given contexts (using first context),
// returns Global tracer if no context is given and GlobalTracer is valid (= non nil, not stopped)
func ExtractTracer(ctx []context.Context) tracer.Tracer {
	if len(ctx) == 0 {
		if tracer.GlobalTracer == nil || tracer.GlobalTracer.Stopped() {
			return nil
		}
		return tracer.GlobalTracer
	}
	rawValue := ctx[0].Value(tracerKeyValue)
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

// MergeTracerContext merges the provided tracer context with the given context.
func MergeTracerContext(ctx context.Context, tracerCtx context.Context) context.Context {
	tracer := ExtractTracer([]context.Context{tracerCtx})
	if tracer != nil {
		return context.WithValue(ctx, tracerKeyValue, tracer)
	}
	return ctx
}
