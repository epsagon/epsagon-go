package epsagongrpc

import (
	"context"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"strconv"
)

func addTraceIdToEventFromContext(ctx context.Context, event *protocol.Event, debugMode bool) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		if debugMode {
			log.Printf("EPSAGON DEBUG Couldn't extract metadata from context: %+v\n", ctx)
		}
		return
	}

	traceIDs := md.Get(EPSAGON_TRACEID_HEADER_KEY)

	if len(traceIDs) == 1 {
		event.Resource.Metadata["epsagon_id"] = traceIDs[0]
	} else {
		if debugMode {
			log.Printf("EPSAGON DEBUG Couldn't extract TraceID from metadata: %+v\n", md)
		}
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

		defer postGRPCRunner(wrapper)

		addTraceIdToEventFromContext(ctx, Event, wrapperTracer.GetConfig().Debug)

		defer wrapperTracer.AddEvent(Event)

		newContext := epsagon.ContextWithTracer(wrapperTracer, ctx)
		wrapperResponse := wrapper.Call(newContext, req)

		resp := wrapperResponse[0].Elem()

		duration := tracer.GetTimestamp() - Event.StartTime

		Event.Duration = duration

		extractGRPCServerRequest(Event.Resource, ctx, info.FullMethod, req)

		var err error = nil
		if !wrapperResponse[1].IsNil() {
			Event.ErrorCode = protocol.ErrorCode_ERROR
			err = wrapperResponse[1].Interface().(error)
		}

		if !wrapperTracer.GetConfig().MetadataOnly {
			Event.Resource.Metadata["rpc.status_code"] = strconv.Itoa(int(status.Code(err)))

			respJson, ok := jsoniter.MarshalToString(resp)
			if ok == nil {
				Event.Resource.Metadata["grpc.response.body"] = respJson
			} else {
				log.Printf("EPSAGON DEBUG err while parsing response: %v", resp)
			}
		}

		return resp.Interface().(interface{}), err
	}
}