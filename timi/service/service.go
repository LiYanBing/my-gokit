package service

import (
	"context"

	timi "github.com/liyanbing/my-gokit/timi/grpc"
)

type service struct {
}

func NewTimi() timi.TimiServer {
	return &service{}
}

func (s *service) HelloWorld(ctx context.Context, req *timi.HelloWorldRequest) (*timi.HelloWorldResponse, error) {
	return &timi.HelloWorldResponse{
		Output: "Hello " + req.Input,
	}, nil
}

func (s *service) Ping(ctx context.Context, req *timi.PingRequest) (*timi.PingResponse, error) {
	return &timi.PingResponse{
		Status: "pong " + req.Service,
	}, nil
}
