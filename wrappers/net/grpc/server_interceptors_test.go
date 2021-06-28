package epsagongrpc

import (
	"context"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/epsagon/epsagon-go/wrappers/net/grpc/testapp"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
	"time"
)


// newTestServerAndConn creates a new *grpc.Server and *grpc.ServerConn for use
// in testing. It adds instrumentation to both. If config is nil, then
// instrumentation is not applied to the server. Be sure to Stop() the server
// and Close() the connection when done with them.
func newTestServerAndConn(
	config *epsagon.Config,
	shouldInstrumentClient bool,
	shouldInstrumentServer bool,
) (*grpc.Server, *grpc.ClientConn) {

	var s *grpc.Server

	if shouldInstrumentServer {
		s = grpc.NewServer(
			grpc.UnaryInterceptor(UnaryServerInterceptor(config)),
		)
	} else {
		s = grpc.NewServer()
	}

	testapp.RegisterTestAppServer(s, &testapp.Server{})
	lis := bufconn.Listen(1024 * 1024)

	go func() {
		s.Serve(lis)
	}()

	bufDialer := func(string, time.Duration) (net.Conn, error) {
		return lis.Dial()
	}

	var conn *grpc.ClientConn
	var err error

	if shouldInstrumentClient {
		conn, err = grpc.Dial("bufnet",
			grpc.WithDialer(bufDialer),
			grpc.WithInsecure(),
			grpc.WithBlock(), // create the connection synchronously
			grpc.WithUnaryInterceptor(UnaryClientInterceptor()),
		)
	} else {
		conn, err = grpc.Dial("bufnet",
			grpc.WithDialer(bufDialer),
			grpc.WithInsecure(),
			grpc.WithBlock(), // create the connection synchronously
		)
	}

	if err != nil {
		gomega.Expect(err).To(gomega.BeNil())
	}

	return s, conn
}

func TestUnaryServerInterceptor (t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "gRPC Server Wrapper")
}

var _ = Describe("gRPC Server Wrapper", func ()  {
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

	})
	AfterEach(func () {
		tracer.GlobalTracer = nil
		err := dummyConn.Close()
		if err != nil {
		}
		dummyServer.Stop()
	})

	Context("test request with client intercepted to", func() {
		BeforeEach(func() {
			dummyServer, dummyConn = newTestServerAndConn(config, true, true)
		})

		It("test unary request with client instrumented", func() {
			client := testapp.NewTestAppClient(dummyConn)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			resp, err := client.DoUnaryRequest(ctx, request)
			Expect(err).To(BeNil())
			Expect(resp.Message).To(Equal(response.Message))
			Expect(events).To(HaveLen(3))
			Expect(events[0].ErrorCode).To(Equal(protocol.ErrorCode_OK))
			Expect(events[1].ErrorCode).To(Equal(protocol.ErrorCode_OK))
		})
	})

	Context("test unary requests", func() {
		BeforeEach(func() {
			dummyServer, dummyConn = newTestServerAndConn(config, false, true)
		})
		It("Sending valid request and validate response without errors", func() {
			client := testapp.NewTestAppClient(dummyConn)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			resp, err := client.DoUnaryRequest(ctx, request)
			Expect(err).To(BeNil())
			Expect(resp.Message).To(Equal(response.Message))
			Expect(events).To(HaveLen(2))
			Expect(events[1].ErrorCode).To(Equal(protocol.ErrorCode_OK))
			Expect(events[1].Resource.Metadata["rpc.status_code"]).To(Equal("0"))
		})

		It ("Send request to errored server and validate error", func () {
			client := testapp.NewTestAppClient(dummyConn)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			resp, err := client.DoUnaryRequestWithError(ctx, request)

			Expect(err).ToNot(BeNil())
			Expect(resp).To(BeNil())
			Expect(events).To(HaveLen(2))
			Expect(events[1].ErrorCode).To(Equal(protocol.ErrorCode_ERROR))
			Expect(events[1].Resource.Metadata["rpc.status_code"]).To(Equal("2"))
		})
	})
})
