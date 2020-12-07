package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"sobe-kit/grpc_tool"

	trans "github.com/go-kit/kit/transport/grpc"
	edp "sobe-kit/order/grpc/endpoints"
	order "sobe-kit/order/grpc/pb"
)

func NewClient(opts ...grpc_tool.ClientOption) (order.OrderServer, error) {
	o := grpc_tool.NewClientOptions()
	for _, opt := range opts {
		opt(o)
	}

	grpcOpts, err := gRPCDialOptions(o.ServerName, o.Cert)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial("order:2048", grpcOpts...)
	if err != nil {
		return nil, err
	}

	wrap := wrapMethod(order.ServiceName, conn)
	return &edp.Endpoints{
		HealthEndpoint: wrap("Health", new(order.HealthResponse)),
	}, nil
}

func wrapMethod(serviceName string, conn *grpc.ClientConn) func(method string, reply interface{}) endpoint.Endpoint {
	return func(method string, reply interface{}) endpoint.Endpoint {
		ep := trans.NewClient(
			conn,
			serviceName,
			method,
			encodeGRPCRequest,
			decodeGRPCResponse,
			reply,
			grpc_tool.GetClientOptions(serviceName, method)...).Endpoint()

		return wrapEndpoint(
			serviceName,
			method,
			ep)
	}
}

// consul load balance
func NewClientWithConsul(consulAddr, dataCenter string, opts ...grpc_tool.ClientOption) (order.OrderServer, error) {
	o := grpc_tool.NewClientOptions()
	for _, opt := range opts {
		opt(o)
	}

	consulClient, err := api.NewClient(&api.Config{
		Address:    consulAddr,
		Datacenter: dataCenter,
	})
	if err != nil {
		return nil, err
	}

	grpcOpts, err := gRPCDialOptions(o.ServerName, o.Cert)
	if err != nil {
		return nil, err
	}

	instanter := consul.NewInstancer(consul.NewClient(consulClient), o.Logger, order.ServiceName, o.Tags, true)
	retryWrap := retryEndpoint(instanter, order.ServiceName, o.Logger, grpcOpts...)
	return &edp.Endpoints{
		HealthEndpoint: retryWrap("Health", new(order.HealthResponse)),
	}, nil
}

func retryEndpoint(instanter sd.Instancer, serviceName string, logger log.Logger, opts ...grpc.DialOption) func(method string, reply interface{}) endpoint.Endpoint {
	return func(method string, reply interface{}) endpoint.Endpoint {
		maxRetry, timeout := grpc_tool.GetMethodMaxRetryAndTimeout(grpc_tool.MethodName(serviceName, method))

		return grpc_tool.Retry(
			maxRetry,
			timeout,
			lb.NewRoundRobin(
				sd.NewEndpointer(
					instanter,
					endpointFactory(serviceName, method, reply, opts...),
					logger)))
	}
}

func endpointFactory(serviceName, method string, reply interface{}, opts ...grpc.DialOption) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		conn, closer, err := grpc_tool.Get(instance, opts...)
		if err != nil {
			return nil, nil, err
		}

		ep := trans.NewClient(
			conn,
			serviceName,
			method,
			encodeGRPCRequest,
			decodeGRPCResponse,
			reply,
			grpc_tool.GetClientOptions(serviceName, method)...).Endpoint()
		return wrapEndpoint(serviceName, method, ep), closer, nil
	}
}

func wrapEndpoint(serviceName, method string, e endpoint.Endpoint) endpoint.Endpoint {
	middles := grpc_tool.GetClientMiddleware(serviceName, method)
	for _, mid := range middles {
		e = mid(e)
	}
	return e
}

func encodeGRPCRequest(_ context.Context, request interface{}) (interface{}, error) {
	return request, nil
}

func decodeGRPCResponse(_ context.Context, reply interface{}) (interface{}, error) {
	return reply, nil
}

func gRPCDialOptions(serverName string, cert []byte) ([]grpc.DialOption, error) {
	var opts []grpc.DialOption
	if cert != nil && len(cert) > 0 {
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(cert) {
			return nil, fmt.Errorf("credentials: failed to append certificates")
		}

		cred := credentials.NewTLS(&tls.Config{ServerName: serverName, RootCAs: cp})
		opts = append(opts, grpc.WithTransportCredentials(cred))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	return opts, nil
}
