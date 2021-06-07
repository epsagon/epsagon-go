package testapp

import (
	"context"
)

// Server is a gRPC server.
type Server struct{}

// DoUnaryUnary is a unary request, unary response method.
func (s *Server) DoUnaryRequest(ctx context.Context, msg *UnaryRequest) (*UnaryResponse, error) {
	return  &UnaryResponse{Message: "Test App Server Response"}, nil
}
