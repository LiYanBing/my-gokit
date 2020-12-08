package grpc_tool

import (
	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
)

type Options struct {
	Cert       []byte   // 证书信息
	Tags       []string // 服务tags
	ServerName string   // 证书服务名称
	Trace      opentracing.Tracer
	Logger     log.Logger
	Collect    Collector
}

func NewOptions() *Options {
	return &Options{
		Logger:  log.NewNopLogger(),
		Trace:   opentracing.NoopTracer{},
		Collect: NewNopMetricsCollector(),
	}
}

type Option func(*Options)

func WithCert(cert []byte) Option {
	return func(opt *Options) {
		opt.Cert = cert
	}
}

func WithTags(tags []string) Option {
	return func(opt *Options) {
		opt.Tags = tags
	}
}

func WithLogger(log log.Logger) Option {
	return func(opt *Options) {
		opt.Logger = log
	}
}

func WithServerName(name string) Option {
	return func(opt *Options) {
		opt.ServerName = name
	}
}

func WithTrace(trace opentracing.Tracer) Option {
	return func(opts *Options) {
		opts.Trace = trace
	}
}

func WithCollect(coll Collector) Option {
	return func(opts *Options) {
		opts.Collect = coll
	}
}
