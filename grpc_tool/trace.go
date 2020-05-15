package grpc_tool

import (
	"github.com/opentracing/opentracing-go"

	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
)

func NopZipKinTracer() (t opentracing.Tracer) {
	tracer, _ := zipkin.NewTracer(nil)
	return zipkinot.Wrap(tracer)
}
