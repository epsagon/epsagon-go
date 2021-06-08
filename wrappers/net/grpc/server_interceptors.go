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


// UnaryServerInterceptor returns a new unary server interceptor for OpenTracing.
func UnaryServerInterceptor(config *epsagon.Config) grpc.UnaryServerInterceptor {
	if config == nil {
		config = &epsagon.Config{}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		wrapperTracer := tracer.CreateTracer(&config.Config)
		wrapperTracer.Start()
		defer wrapperTracer.Stop()

		Event := createGRPCEvent(info.FullMethod, "grpc-server")

		defer wrapperTracer.AddEvent(Event)

		resp, err := handler(ctx, req)
		duration := tracer.GetTimestamp() - Event.StartTime

		Event.Duration = duration

		decorateGRPCRequest(Event.Resource, ctx, info.FullMethod, req)

		if err != nil {
			Event.ErrorCode = protocol.ErrorCode_ERROR
		}

		Event.Resource.Metadata["status_code"] = strconv.Itoa(int(status.Code(err)))
		Event.Resource.Metadata["grpc.response.body"] = fmt.Sprintf("%+v" , resp)
		Event.Resource.Metadata["span.kind"] = "server"

		return resp, err
	}
}