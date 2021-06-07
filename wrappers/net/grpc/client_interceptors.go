package epsagongrpc

import (
	"context"
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"strconv"
)


// UnaryClientInterceptor returns a new unary server interceptor for OpenTracing.
func UnaryClientInterceptor(config *epsagon.Config) grpc.UnaryClientInterceptor {
	if config == nil {
		config = &epsagon.Config{}
	}

	return func (ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		wrapperTracer := tracer.CreateTracer(&config.Config)
		wrapperTracer.Start()
		defer wrapperTracer.Stop()

		Event := createGRPCEvent(wrapperTracer, ctx, method, "grpc-client")
		decorateGRPCRequest(Event.Resource, ctx, method, req)

		defer wrapperTracer.AddEvent(Event)
		err := invoker(ctx, method, req, reply, cc, opts...)

		duration := tracer.GetTimestamp() - Event.StartTime
		Event.Duration = duration

		if err != nil {
			Event.ErrorCode = protocol.ErrorCode_ERROR
		}

		Event.Resource.Metadata["status_code"] = strconv.Itoa(int(status.Code(err)))
		Event.Resource.Metadata["grpc.response.body"] = fmt.Sprintf("%+v" , reply)

		return err
	}
}
