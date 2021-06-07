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
	RunSpecs(t, "Gin Wrapper")
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

	BeforeEach(func(t *testing.T) {
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
		response = &testapp.UnaryResponse{}

		dummyServer, dummyConn = newTestServerAndConn(t, config)
	})

	AfterEach(func (t *testing.T) {
		tracer.GlobalTracer = nil
		err := dummyConn.Close()
		if err != nil {
			t.Fatal("failure to finish Connection", err)
		}
		dummyServer.Stop()
	})

	Context("test sanity", func() {
		It("Sending valid request and validate response without errors", func() {
			client := testapp.NewTestAppClient(dummyConn)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			resp, err := client.DoUnaryRequest(ctx, request)
			Expect(err).To(Not(Equal(nil)))
			Expect(resp.Message).To(Equal(response))

		})
	})
})