package transport

import (
	"context"

	"github.com/go-kit/kit/transport/grpc"
	"sobe-kit/grpc_tool"
	"sobe-kit/order/grpc/endpoints"

	order "sobe-kit/order/grpc/pb"
)

type gRPCServer struct {
	healthHandler grpc.Handler
}

func (g *gRPCServer) Health(ctx context.Context, req *order.HealthRequest) (*order.HealthResponse, error) {
	_, ret, err := g.healthHandler.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return ret.(*order.HealthResponse), nil
}

func NewGRPCServer(service order.OrderServer) order.OrderServer {
	eps := endpoints.WrapEndpoints(order.ServiceName, service)
	return &gRPCServer{
		healthHandler: grpc.NewServer(
			eps.HealthEndpoint,
			DecodeRequestFunc,
			EncodeResponseFunc,
			grpc_tool.GetServerOptions(order.ServiceName, "Health")...),
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
