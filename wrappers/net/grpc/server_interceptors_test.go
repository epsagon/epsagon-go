package epsagongrpc

import (
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/wrappers/net/grpc/testapp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
	"time"
)

// newTestServerAndConn creates a new *grpc.Server and *grpc.ClientConn for use
// in testing. It adds instrumentation to both. If app is nil, then
// instrumentation is not applied to the server. Be sure to Stop() the server
// and Close() the connection when done with them.
func newTestServerAndConn(t *testing.T, config *epsagon.Config) (*grpc.Server, *grpc.ClientConn) {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(UnaryServerInterceptor(config)),
	)

	testapp.RegisterTestAppServer(s, &testapp.Server{})
	lis := bufconn.Listen(1024 * 1024)

	go func() {
		s.Serve(lis)
	}()

	bufDialer := func(string, time.Duration) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.Dial("bufnet",
		grpc.WithDialer(bufDialer),
		grpc.WithInsecure(),
		grpc.WithBlock(), // create the connection synchronously
		grpc.WithUnaryInterceptor(UnaryClientInterceptor(config)),
	)
	if err != nil {
		t.Fatal("failure to create ClientConn", err)
	}

	return s, conn
}
