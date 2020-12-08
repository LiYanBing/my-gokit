package grpc_tool

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	"google.golang.org/grpc/status"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type Collector interface {
	Collect(serviceName, method string) endpoint.Middleware
}

type MetricsCollector struct {
	ReqCounter metrics.Counter   // counter
	ReqLatency metrics.Histogram // histogram
}

func (col *MetricsCollector) Collect(serviceName, method string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (resp interface{}, err error) {
			defer func(begin time.Time) {
				lvs := []string{"method", method, "code", col.parseError(err), "service", serviceName, "hostname", hostname}
				col.ReqCounter.With(lvs...).Add(1)
				col.ReqLatency.With(lvs...).Observe(time.Since(begin).Seconds())
			}(time.Now())
			resp, err = next(ctx, request)
			return
		}
	}
}

func (col *MetricsCollector) parseError(err error) string {
	if err == nil {
		return "0"
	}
	code := uint32(status.Code(err))
	return fmt.Sprintf("%v", code)

}

func NewMetricsCollector() *MetricsCollector {
	reqCounter := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Name: "request_counter",
		Help: "Number of requests received.",
	}, []string{"method", "code", "service", "hostname"})

	requestLatency := kitprometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Name:    "request_duration",
		Help:    "Total duration of requests . not include writing on wire",
		Buckets: []float64{.1, 1, 10, 30, 50, 100, 300, 500, 1000, 3000},
	}, []string{"method", "code", "service", "hostname"})

	return &MetricsCollector{
		ReqCounter: reqCounter,
		ReqLatency: requestLatency,
	}
}

func NewNopMetricsCollector() *nopMetricsCollector {
	return &nopMetricsCollector{}
}

type nopMetricsCollector struct{}

func (nop nopMetricsCollector) Collect(serviceName, method string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			return next(ctx, request)
		}
	}
}

var hostname string

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		panic(err)
	}
}

func NewExposer(addr string) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(addr, nil); err != nil {
			panic(err)
		}
	}()
}
