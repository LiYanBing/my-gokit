package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"sobe-kit/grpc_tool"

	order "sobe-kit/order/grpc/pb"
)

type Endpoints struct {
	HealthEndpoint endpoint.Endpoint
}

func (e *Endpoints) Health(ctx context.Context, req *order.HealthRequest) (*order.HealthResponse, error) {
	ret, err := e.HealthEndpoint(ctx, req)
	if err != nil {
		return nil, err
	}

	return ret.(*order.HealthResponse), nil
}

func wrapHealth(service order.OrderServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*order.HealthRequest)
		return service.Health(ctx, req)
	}
}

func wrapEndpoint(service, method string, endpoint endpoint.Endpoint) endpoint.Endpoint {
	wrappers := grpc_tool.GetEndpointMiddleware(service, method)
	for _, wrap := range wrappers {
		endpoint = wrap(endpoint)
	}
	return endpoint
}

func WrapEndpoints(serviceName string, service order.OrderServer) *Endpoints {
	return &Endpoints{
		HealthEndpoint: wrapEndpoint(serviceName, "Health", wrapHealth(service)),
	}
}
