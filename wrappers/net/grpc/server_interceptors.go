package grpc

import (
	"context"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/google/uuid"

	"google.golang.org/grpc"
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

		Event := createGRPCEvent(wrapperTracer, ctx, info.FullMethod)
		defer wrapperTracer.AddEvent(Event)

		resp, err := handler(ctx, req)
		duration := tracer.GetTimestamp() - Event.StartTime

		Event.Duration = duration

		return resp, err
	}
}

func updateResponseData(resp interface{}, resource *protocol.Resource, metadataOnly bool) {

}

func createGRPCEvent(wrapperTacer tracer.Tracer, ctx context.Context, method string) *protocol.Event {
	errorcode := protocol.ErrorCode_OK

	return &protocol.Event{
		Id:        "grpc-server" + uuid.New().String(),
		Origin:    "grpc.Server",
		StartTime: tracer.GetTimestamp(),
		ErrorCode: errorcode,
		Resource: &protocol.Resource{
			Name:      "url",
			Type:      "grpc",
			Operation: method,
			Metadata:  map[string]string{},
		},
	}
}