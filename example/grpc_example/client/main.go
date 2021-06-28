/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"github.com/epsagon/epsagon-go/epsagon"
	"log"
	"os"
	"time"

	epsagongrpc "github.com/epsagon/epsagon-go/wrappers/net/grpc"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func setupClient() (int, error) {

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithUnaryInterceptor(epsagongrpc.UnaryClientInterceptor()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())

	return 0, nil
}


func main() {
	// Set up a connection to the server.
	config := epsagon.NewTracerConfig(
		"grcp-client-wrapper-test", "05b05129-e34e-40ea-873a-85c45a6d5b3f",
	)

	config.Debug = true
	config.CollectorURL = "http://dev.tc.epsagon.com/"
	config.MetadataOnly = false
	config.SendTimeout = "10s"
	epsagon.GoWrapper(config, setupClient)()
}
