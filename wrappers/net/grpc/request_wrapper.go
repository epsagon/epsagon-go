package grpc

import (
	"github.com/epsagon/epsagon-go/tracer"
)

const EPSAGON_TRACEID_HEADER_KEY = "epsagon-trace-id"
const EPSAGON_DOMAIN = "epsagon.com"


// OpNameFunc is a func that allows custom operation names instead of the gRPC method.
type OpNameFunc func(method string) string


// UnaryHandlerTracer is Epsagon's wrapper for grpc Unary request handlers
type UnaryHandlerTracer struct {

	// MetadataOnly flag overriding the configuration
	MetadataOnly bool
	traceHeaderName         string
	unaryRequestHandlerFunc interface{}
	opNameFunc              OpNameFunc

	tracer       tracer.Tracer
}

type Arg func(wrapper *UnaryHandlerTracer)
