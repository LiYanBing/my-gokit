package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/liyanbing/my-gokit/grpc_tool"

	timi "github.com/liyanbing/my-gokit/example/timi/grpc"
)

type Endpoints struct {
	HelloWorldEndpoint endpoint.Endpoint
	PingEndpoint       endpoint.Endpoint
}

func (e *Endpoints) HelloWorld(ctx context.Context, req *timi.HelloWorldRequest) (*timi.HelloWorldResponse, error) {
	ret, err := e.HelloWorldEndpoint(ctx, req)
	if err != nil {
		return nil, err
	}

	return ret.(*timi.HelloWorldResponse), nil
}

func (e *Endpoints) Ping(ctx context.Context, req *timi.PingRequest) (*timi.PingResponse, error) {
	ret, err := e.PingEndpoint(ctx, req)
	if err != nil {
		return nil, err
	}

	return ret.(*timi.PingResponse), nil
}

func wrapHelloWorld(service timi.TimiServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*timi.HelloWorldRequest)
		return service.HelloWorld(ctx, req)
	}
}

func wrapPing(service timi.TimiServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*timi.PingRequest)
		return service.Ping(ctx, req)
	}
}

func wrapEndpoint(service, method string, endpoint endpoint.Endpoint) endpoint.Endpoint {
	wrappers := grpc_tool.GetEndpointMiddleware(service, method)
	for _, wrap := range wrappers {
		endpoint = wrap(endpoint)
	}
	return endpoint
}

func WrapEndpoints(serviceName string, service timi.TimiServer) *Endpoints {
	return &Endpoints{
		HelloWorldEndpoint: wrapEndpoint(serviceName, "HelloWorld", wrapHelloWorld(service)),
		PingEndpoint:       wrapEndpoint(serviceName, "Ping", wrapPing(service)),
	}
}
