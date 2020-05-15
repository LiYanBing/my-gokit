package grpc_tool

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd/lb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type retryGRPCKey int

const (
	noRetryGRPCKey retryGRPCKey = iota
)

func Retry(max int, timeout time.Duration, b lb.Balancer) endpoint.Endpoint {
	return retryWithFinalRawGRPCError(max, timeout, b)
}

func NoRetryGRPCContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, noRetryGRPCKey, struct{}{})
}

func isNoRetryGRPC(ctx context.Context) bool {
	return ctx.Value(noRetryGRPCKey) != nil
}

func retryWithFinalRawGRPCError(max int, timeout time.Duration, b lb.Balancer) endpoint.Endpoint {
	if timeout > 0 {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if isNoRetryGRPC(ctx) {
				var e endpoint.Endpoint
				e, err = b.Endpoint()
				if err != nil {
					return
				}
				return e(ctx, request)
			}

			response, err = lb.RetryWithCallback(timeout, b, maxRetries(max))(ctx, request)
			if err != nil {
				if e, ok := err.(lb.RetryError); ok {
					err = e.Final
				}
			}
			return
		}
	}

	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		e, err := b.Endpoint()
		if err != nil {
			return
		}
		return e(ctx, request)
	}
}

func shouldRetryError(err error) bool {
	if s, ok := status.FromError(err); ok {
		switch s.Code() {
		case codes.DeadlineExceeded:
			fallthrough
		case codes.NotFound:
			fallthrough
		case codes.AlreadyExists:
			fallthrough
		case codes.ResourceExhausted:
			fallthrough
		case codes.FailedPrecondition:
			fallthrough
		case codes.Aborted:
			fallthrough
		case codes.Unimplemented:
			fallthrough
		case codes.Unavailable:
			fallthrough
		case codes.DataLoss:
			return true
		default:
			// others are reserved for service use
			return false
		}
	}
	return false
}

func maxRetries(max int) lb.Callback {
	return func(n int, err error) (keepTrying bool, replacement error) {
		if shouldRetryError(err) {
			return n < max, nil
		} else {
			return false, nil
		}
	}
}
