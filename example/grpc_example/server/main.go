// Package main implements a server for Greeter service.
package main

import (
	"context"
	"log"
	"net"

	epsagongrpc "github.com/epsagon/epsagon-go/wrappers/net/grpc"
	epsagon "github.com/epsagon/epsagon-go/epsagon"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}


func main() {
	config := epsagon.NewTracerConfig(
		"grcp-server-wrapper-test", "",
	)
	config.MetadataOnly = false

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(epsagongrpc.UnaryServerInterceptor(config)),
	)

	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
