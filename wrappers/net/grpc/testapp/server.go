package testapp

import (
	"context"
	"errors"
)

// Server is a gRPC server.
type Server struct{}

// DoUnaryUnary is a unary request, unary response method.
func (s *Server) DoUnaryRequest(ctx context.Context, msg *UnaryRequest) (*UnaryResponse, error) {
	return  &UnaryResponse{Message: "Test App Server Response"}, nil
}

func (s *Server) DoUnaryRequestWithError(ctx context.Context, msg *UnaryRequest) (*UnaryResponse, error) {
	return &UnaryResponse{Message: "Test App Server Response"}, errors.New("New Error")
}
