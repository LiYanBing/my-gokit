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
	"github.com/liyanbing/my-gokit/grpc_tool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	trans "github.com/go-kit/kit/transport/grpc"
	timi "github.com/liyanbing/my-gokit/timi/grpc"
	edp "github.com/liyanbing/my-gokit/timi/grpc/endpoints"
)

func NewClient(addr, serverName string, cert []byte) (timi.TimiServer, error) {
	opts, err := gRPCDialOptions(serverName, cert)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	wrap := wrapMethod(serverName, conn)
	return &edp.Endpoints{
		HelloWorldEndpoint: wrap("HelloWorld", new(timi.HelloWorldResponse)),
		PingEndpoint:       wrap("Ping", new(timi.PingResponse)),
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
func NewClientWithConsul(consulAddr, dataCenter, serverName string, cert []byte, tags []string, logger log.Logger) (timi.TimiServer, error) {
	consulClient, err := api.NewClient(&api.Config{
		Address:    consulAddr,
		Datacenter: dataCenter,
	})
	if err != nil {
		return nil, err
	}

	opts, err := gRPCDialOptions(serverName, cert)
	if err != nil {
		return nil, err
	}

	instanter := consul.NewInstancer(consul.NewClient(consulClient), logger, timi.ServiceName, tags, true)
	retryWrap := retryEndpoint(instanter, serverName, logger, opts...)
	return &edp.Endpoints{
		HelloWorldEndpoint: retryWrap("HelloWorld", new(timi.HelloWorldResponse)),
		PingEndpoint:       retryWrap("Ping", new(timi.PingResponse)),
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
