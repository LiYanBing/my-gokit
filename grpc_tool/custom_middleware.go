package grpc_tool

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport/grpc"
)

/**
 *********** endpoint middleware ***********
 */
type Wrapper func(service, method string) endpoint.Middleware

var (
	endpointMiddleware []Wrapper
)

func RegisterEndpointMiddlewareChain(wrapper ...Wrapper) {
	if wrapper == nil {
		return
	}

	endpointMiddleware = append(endpointMiddleware, wrapper...)
}

func GetEndpointMiddleware(service, method string) []endpoint.Middleware {
	middles := make([]endpoint.Middleware, 0, len(endpointMiddleware))

	for _, wrapper := range endpointMiddleware {
		middles = append(middles, wrapper(service, method))
	}

	return middles
}

/**
 *********** grpc server and client options ***********
 */
var (
	clientOptions []ClientOptionFunc
	serverOptions []ServerOptionFunc
)

type ClientOptionFunc func(serverName, method string) grpc.ClientOption

type ServerOptionFunc func(serverName, method string) grpc.ServerOption

func RegisterClientOptionFunc(options ...ClientOptionFunc) {
	clientOptions = append(clientOptions, options...)
}

func GetClientOptions(serverName, method string) []grpc.ClientOption {
	options := make([]grpc.ClientOption, 0, len(clientOptions))
	for _, opt := range clientOptions {
		options = append(options, opt(serverName, method))
	}

	return options
}

func RegisterServerOptionFunc(options ...ServerOptionFunc) {
	serverOptions = append(serverOptions, options...)
}

func GetServerOptions(serverName, method string) []grpc.ServerOption {
	options := make([]grpc.ServerOption, 0, len(serverOptions))
	for _, opt := range serverOptions {
		options = append(options, opt(serverName, method))
	}

	return options
}

/**
 *********** grpc client middleware ***********
 */
var (
	clientMiddleware []Wrapper
)

func RegisterClientMiddleware(wrapper ...Wrapper) {
	if wrapper == nil {
		return
	}

	clientMiddleware = append(clientMiddleware, wrapper...)
}

func GetClientMiddleware(service, method string) []endpoint.Middleware {
	middles := make([]endpoint.Middleware, 0, len(endpointMiddleware))

	for _, wrapper := range clientMiddleware {
		middles = append(middles, wrapper(service, method))
	}

	return middles
}

/**
 ************ grpc client max retry and timeout ***********
 */

type redirect struct {
	maxRetry int
	timeout  time.Duration
}

var (
	MaxRetry = 3
	Timeout  = time.Second * 5
	mapping  map[string]*redirect
)

func MethodName(serviceName, method string) string {
	return fmt.Sprintf("%v.%v", serviceName, method)
}

func RedirectRetryAndTimeout(name string, max int, t time.Duration) {
	if mapping == nil {
		mapping = make(map[string]*redirect)
	}

	mapping[name] = &redirect{
		maxRetry: max,
		timeout:  t,
	}
}

func GetMethodMaxRetryAndTimeout(name string) (int, time.Duration) {
	if value, ok := mapping[name]; ok {
		return value.maxRetry, value.timeout
	}

	return MaxRetry, Timeout
}
