package grpc_tool

import "github.com/go-kit/kit/log"

type ClientOptions struct {
	Cert       []byte     // 证书信息
	Tags       []string   // 服务tags
	Logger     log.Logger // 日志
	ServerName string     // 证书服务名称
}

func NewClientOptions() *ClientOptions {
	return &ClientOptions{
		Logger: log.NewNopLogger(),
	}
}

type ClientOption func(opt *ClientOptions)

func WithCert(cert []byte) ClientOption {
	return func(opt *ClientOptions) {
		opt.Cert = cert
	}
}

func WithTags(tags []string) ClientOption {
	return func(opt *ClientOptions) {
		opt.Tags = tags
	}
}

func WithLogger(log log.Logger) ClientOption {
	return func(opt *ClientOptions) {
		opt.Logger = log
	}
}

func WithServerName(name string) ClientOption {
	return func(opt *ClientOptions) {
		opt.ServerName = name
	}
}
