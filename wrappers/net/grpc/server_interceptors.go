package epsagongrpc

import (
	"context"
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"strconv"
)

func addTraceIdToEventFromContext(ctx context.Context, event *protocol.Event) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		log.Printf("EPSAGON DEBUG Couldn't extract metadata from context: %+v\n", ctx)
	}

	traceIDs := md.Get(EPSAGON_TRACEID_HEADER_KEY)
	if len(traceIDs) == 1 {
		event.Resource.Metadata[tracer.EpsagonGRPCraceIDKey] = traceIDs[0]
	} else {
		log.Printf("EPSAGON DEBUG Couldn't extract TraceID from metadata: %+v\n", md)
	}
}

// UnaryServerInterceptor returns a new unary server interceptor for OpenTracing.
func UnaryServerInterceptor(config *epsagon.Config) grpc.UnaryServerInterceptor {
	if config == nil {
		config = &epsagon.Config{}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		wrapperTracer := tracer.CreateTracer(&config.Config)
		wrapperTracer.Start()
		defer wrapperTracer.Stop()

		Event := createGRPCEvent("trigger", info.FullMethod, "grpc-server")

		wrapper := epsagon.WrapGenericFunction(
			handler, config, wrapperTracer, false, info.FullMethod,
		)

		defer decoratePostGRPCRunner(wrapper)

		addTraceIdToEventFromContext(ctx, Event)

		defer wrapperTracer.AddEvent(Event)

		wrapperResponse := wrapper.Call(ctx, req)

		resp := wrapperResponse[0].Elem()

		duration := tracer.GetTimestamp() - Event.StartTime

		Event.Duration = duration

		decorateGRPCRequest(Event.Resource, ctx, info.FullMethod, req)

		var err error = nil
		if !wrapperResponse[1].IsNil() {
			Event.ErrorCode = protocol.ErrorCode_ERROR
			err = wrapperResponse[1].Interface().(error)
		}

		Event.Resource.Metadata["status_code"] = strconv.Itoa(int(status.Code(err)))
		Event.Resource.Metadata["grpc.response.body"] = fmt.Sprintf("%+v" , resp)
		Event.Resource.Metadata["span.kind"] = "server"

		return resp.Interface().(interface{}), err
	}
}