package trace

import (
	"context"
	"fmt"
	"io"
	"strings"

	"sobe-kit/grpc_tool"

	"github.com/go-kit/kit/endpoint"
	"github.com/opentracing/opentracing-go"

	opentracinglog "github.com/opentracing/opentracing-go/log"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport"
)

func NopZipKinTracer() (t opentracing.Tracer) {
	tracer, _ := zipkin.NewTracer(nil)
	return zipkinot.Wrap(tracer)
}

func StartSpanFromContext(ctx context.Context, tracer opentracing.Tracer, name string, opts ...opentracing.StartSpanOption) (context.Context, opentracing.Span, error) {
	md, ok := FromContext(ctx)
	if !ok {
		md = make(Metadata)
	}

	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		opts = append(opts, opentracing.ChildOf(parentSpan.Context()))
	} else if spanCtx, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(md)); err == nil {
		opts = append(opts, opentracing.ChildOf(spanCtx))
	}

	nmd := make(Metadata, 1)

	sp := tracer.StartSpan(name, opts...)

	if err := sp.Tracer().Inject(sp.Context(), opentracing.TextMap, opentracing.TextMapCarrier(nmd)); err != nil {
		return nil, nil, err
	}

	for k, v := range nmd {
		md.Set(strings.Title(k), v)
	}

	ctx = opentracing.ContextWithSpan(ctx, sp)
	ctx = NewContext(ctx, md)
	return ctx, sp, nil
}

func Trace(trace opentracing.Tracer) grpc_tool.Wrapper {
	return func(service, method string) endpoint.Middleware {
		return func(next endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, request interface{}) (res interface{}, err error) {
				if trace == nil {
					trace = opentracing.GlobalTracer()
				}

				name := fmt.Sprintf("%s/%s", service, method)
				ctx, span, err := StartSpanFromContext(ctx, trace, name)
				if err != nil {
					return nil, err
				}
				defer span.Finish()

				if res, err = next(ctx, request); err != nil {
					span.LogFields(opentracinglog.String("error", err.Error()))
					span.SetTag("error", true)
				}
				return res, err
			}
		}
	}
}

func NewJaegerTracer(service, url string) (opentracing.Tracer, io.Closer) {
	sender := transport.NewHTTPTransport(url)
	tracer, closer := jaeger.NewTracer(
		service,
		jaeger.NewConstSampler(true),
		jaeger.NewRemoteReporter(sender))
	return tracer, closer
}
