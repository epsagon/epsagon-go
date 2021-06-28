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
	"strconv"
)

const EPSAGON_TRACEID_HEADER_KEY = "epsagon-trace-id"

func InjectEpsagonTracerContextID(ctx context.Context, event *protocol.Event) context.Context {
	traceID := epsagon.GenerateEpsagonTraceID()

	addTraceIdToEvent(traceID, event)

	return metadata.AppendToOutgoingContext(ctx, EPSAGON_TRACEID_HEADER_KEY, traceID)
}

func addTraceIdToEvent(traceID string, event *protocol.Event) {
	event.Resource.Metadata[tracer.EpsagonGRPCTraceIDKey] = traceID
}


// UnaryClientInterceptor returns a new unary server interceptor for OpenTracing.
func UnaryClientInterceptor(args ...context.Context) grpc.UnaryClientInterceptor {

	wrapperTracer := epsagon.ExtractTracer(args)
	return func (ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		Event := createGRPCEvent("grpc.client", method, "grpc-client")
		ctx = InjectEpsagonTracerContextID(ctx, Event)


		defer wrapperTracer.AddEvent(Event)
		err := invoker(ctx, method, req, reply, cc, opts...)

		defer extractGRPCClientRequest(Event.Resource, method, req, cc.Target())

		duration := tracer.GetTimestamp() - Event.StartTime
		Event.Duration = duration

		Event.Resource.Metadata["rpc.status_code"] = strconv.Itoa(int(status.Code(err)))

		if err != nil {
			Event.ErrorCode = protocol.ErrorCode_ERROR
			return err
		}

		replyJson, err := jsoniter.MarshalToString(reply)
		if err == nil {
			Event.Resource.Metadata["grpc.response.body"] = replyJson
		}

		return err
	}
}
