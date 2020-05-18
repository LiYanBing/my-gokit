package transport

import (
	"context"

	"github.com/go-kit/kit/transport/grpc"
	"github.com/liyanbing/my-gokit/grpc_tool"
	"github.com/liyanbing/my-gokit/timi/grpc/endpoints"

	timi "github.com/liyanbing/my-gokit/timi/grpc"
)

type gRPCServer struct {
	helloWorldHandler grpc.Handler
	pingHandler       grpc.Handler
}

func (g *gRPCServer) HelloWorld(ctx context.Context, req *timi.HelloWorldRequest) (*timi.HelloWorldResponse, error) {
	_, ret, err := g.helloWorldHandler.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return ret.(*timi.HelloWorldResponse), nil
}

func (g *gRPCServer) Ping(ctx context.Context, req *timi.PingRequest) (*timi.PingResponse, error) {
	_, ret, err := g.pingHandler.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return ret.(*timi.PingResponse), nil
}

func NewGRPCServer(service timi.TimiServer) timi.TimiServer {
	eps := endpoints.WrapEndpoints(timi.ServiceName, service)
	return &gRPCServer{
		helloWorldHandler: grpc.NewServer(
			eps.HelloWorldEndpoint,
			DecodeRequestFunc,
			EncodeResponseFunc,
			grpc_tool.GetServerOptions(timi.ServiceName, "HelloWorld")...),

		pingHandler: grpc.NewServer(
			eps.PingEndpoint,
			DecodeRequestFunc,
			EncodeResponseFunc,
			grpc_tool.GetServerOptions(timi.ServiceName, "Ping")...),
	}
}

func DecodeRequestFunc(_ context.Context, req interface{}) (request interface{}, err error) {
	request = req
	return
}

func EncodeResponseFunc(_ context.Context, resp interface{}) (response interface{}, err error) {
	response = resp
	return
}
