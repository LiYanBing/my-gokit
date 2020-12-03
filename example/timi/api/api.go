package timi

import (
	"context"

	timi "sobe-kit/example/timi/grpc"
)

type Timi interface {
	// say hello world
	HelloWorld(context.Context, *timi.HelloWorldRequest) (*timi.HelloWorldResponse, error)
	// Ping
	Ping(context.Context, *timi.PingRequest) (*timi.PingResponse, error)
}
