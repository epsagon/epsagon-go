package epsagongrpc

import (
	"context"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/epsagon/epsagon-go/wrappers/net/grpc/testapp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestUnaryClientInterceptor (t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "gRPC Client Wrapper")
}

func verifyTraceIDExists(event *protocol.Event) {
	traceID, ok := event.Resource.Metadata[tracer.EpsagonGRPCTraceIDKey]
	Expect(ok).To(BeTrue())
	Expect(traceID).To(Not(BeZero()))
}


var _ = Describe("gRPC Client Wrapper", func ()  {
	var (
		events         []*protocol.Event
		exceptions     []*protocol.Exception
		request        *testapp.UnaryRequest
		response 	   *testapp.UnaryResponse
		config         *epsagon.Config
		dummyServer	   *grpc.Server
		dummyConn	   *grpc.ClientConn
	)
	BeforeEach(func() {
		config = &epsagon.Config{Config: tracer.Config{
			Disable:  true,
			TestMode: true,
		}}

		events = make([]*protocol.Event, 0)
		exceptions = make([]*protocol.Exception, 0)

		tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
			Events:     &events,
			Exceptions: &exceptions,
			Labels:     make(map[string]interface{}),
			Config:     &config.Config,
		}

		request = &testapp.UnaryRequest{Message: "Test Message"}
		response = &testapp.UnaryResponse{Message: "Test App Server Response"}

		dummyServer, dummyConn = newTestServerAndConn(config, true, false)
	})
	AfterEach(func () {
		tracer.GlobalTracer = nil
		err := dummyConn.Close()
		if err != nil {
		}
		dummyServer.Stop()
	})

	Context("test unary requests", func() {
		It("Sending valid request and validate response without errors", func() {
			client := testapp.NewTestAppClient(dummyConn)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			resp, err := client.DoUnaryRequest(ctx, request)
			Expect(err).To(BeNil())
			Expect(resp.Message).To(Equal(response.Message))
			Expect(events).To(HaveLen(1))
			Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
			Expect(events[0].Resource.Metadata["rpc.status_code"]).To(Equal("0"))
			verifyTraceIDExists(events[0])
		})

		It ("Send request to errored server and validate error", func () {
			client := testapp.NewTestAppClient(dummyConn)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			resp, err := client.DoUnaryRequestWithError(ctx, request)

			Expect(err).ToNot(BeNil())
			Expect(resp).To(BeNil())
			Expect(events).To(HaveLen(1))
			Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_ERROR))
			Expect(events[0].Resource.Metadata["rpc.status_code"]).To(Equal("2"))
			verifyTraceIDExists(events[0])
		})
	})
})