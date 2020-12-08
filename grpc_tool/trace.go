package grpc_tool

import (
	"io"

	"github.com/opentracing/opentracing-go"

	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport"
)

func NopZipKinTracer() (t opentracing.Tracer) {
	tracer, _ := zipkin.NewTracer(nil)
	return zipkinot.Wrap(tracer)
}

func NewJaegerTracer(service, url string) (opentracing.Tracer, io.Closer) {
	sender := transport.NewHTTPTransport(url)
	tracer, closer := jaeger.NewTracer(
		service,
		jaeger.NewConstSampler(true),
		jaeger.NewRemoteReporter(sender))
	return tracer, closer
}
