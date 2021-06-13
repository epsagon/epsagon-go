package epsagongrpc

import (
	"context"
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"strconv"
	"strings"
)


const EPSAGON_TRACEID_HEADER_KEY = "epsagon-trace-id"

func generateRandomUUID() string {
	uuid, err := uuid.NewRandom()
	if err != nil {
		panic("failed to generate random UUID")
	}
	return strings.ReplaceAll(uuid.String(), "-", "")
}

func generateEpsagonTraceID() string {
	traceID := generateRandomUUID()
	spanID := generateRandomUUID()[:16]
	parentSpanID := generateRandomUUID()[:16]
	return fmt.Sprintf("%s:%s:%s:1", traceID, spanID, parentSpanID)
}

func InjectEpsagonTracerContextID(ctx context.Context, event *protocol.Event) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		log.Printf("EPSAGON DEBUG Couldn't inject TraceID to context: %+v\n", ctx)
		md = metadata.New(nil)
	}

	traceID := generateEpsagonTraceID()

	md.Set(EPSAGON_TRACEID_HEADER_KEY, traceID)
	addTraceIdToEvent(traceID, event)

	ctx = metadata.NewOutgoingContext(ctx, md)
}

func addTraceIdToEvent(traceID string, event *protocol.Event) {
	event.Resource.Metadata[tracer.EpsagonGRPCraceIDKey] = traceID
}


// UnaryClientInterceptor returns a new unary server interceptor for OpenTracing.
func UnaryClientInterceptor(config *epsagon.Config) grpc.UnaryClientInterceptor {
	if config == nil {
		config = &epsagon.Config{}
	}

	return func (ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		wrapperTracer := tracer.CreateTracer(&config.Config)
		wrapperTracer.Start()
		defer wrapperTracer.Stop()

		Event := createGRPCEvent("runner", method, "grpc-client")
		InjectEpsagonTracerContextID(ctx, Event)

		decorateGRPCRequest(Event.Resource, ctx, method, req)

		defer wrapperTracer.AddEvent(Event)
		err := invoker(ctx, method, req, reply, cc, opts...)

		duration := tracer.GetTimestamp() - Event.StartTime
		Event.Duration = duration

		Event.Resource.Metadata["status_code"] = strconv.Itoa(int(status.Code(err)))
		Event.Resource.Metadata["span.kind"] = "client"

		if err != nil {
			Event.ErrorCode = protocol.ErrorCode_ERROR
			return err
		}

		Event.Resource.Metadata["grpc.response.body"] = fmt.Sprintf("%+v" , reply)

		return err
	}
}
