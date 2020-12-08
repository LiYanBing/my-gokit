package gokit_tool

const (
	protoTemplate = `syntax = "proto3";

option go_package = "pb;{{.PkgName}}";

service {{.ServiceName}} {
	rpc Health (HealthRequest) returns (HealthResponse) {}
}

message HealthRequest {
	string service = 1;
}

message HealthResponse {
	int64 status = 1;
}
`

	constantTemplate = `package {{.PkgName}} 

var (
	ServiceName = _{{.ServiceName}}_serviceDesc.ServiceName
)`
	apiTemplate = `
package {{.PkgName}}

import (
	"context"

	{{.PkgName}} "{{.ImportPath}}"
)

type {{.ServiceName}} interface {
	{{- range .Methods}}
		{{.Doc}}
		{{- .Name}}(context.Context, *{{$.PkgName}}.{{.RequestName}}) (*{{$.PkgName}}.{{.ResponseName}}, error)
	{{- end}}
}`

	endpointsTemplate = `
package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/tracing/opentracing"
	"sobe-kit/grpc_tool"

	{{.PkgName}} "{{.ImportPath}}"
)

type Endpoints struct {
{{- range .Methods}}
	{{.Name}}Endpoint endpoint.Endpoint
{{- end}}
}

{{range .Methods}}
func (e *Endpoints) {{.Name}}(ctx context.Context, req *{{$.PkgName}}.{{.RequestName}}) (*{{$.PkgName}}.{{.ResponseName}}, error) {
	ret, err := e.{{.Name}}Endpoint(ctx, req)
	if err != nil {
		return nil, err
	}

	return ret.(*{{$.PkgName}}.{{.ResponseName}}), nil
}
{{end}}

{{range .Methods}}
func wrap{{.Name}}(service {{$.PkgName}}.{{$.ServiceName}}Server) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*{{$.PkgName}}.{{.RequestName}})
		return service.{{.Name}}(ctx, req)
	}
}
{{end}}

func wrapEndpoint(service, method string, endpoint endpoint.Endpoint, opt *grpc_tool.Options) endpoint.Endpoint {
	endpoint = grpc_tool.Recover(service, method)(endpoint)
	endpoint = opt.Collect.Collect(service, method)(endpoint)
	endpoint = opentracing.TraceServer(opt.Trace, method)(endpoint)

	wrappers := grpc_tool.GetEndpointMiddleware(service, method)
	for _, wrap := range wrappers {
		endpoint = wrap(endpoint)
	}
	return endpoint
}

func WrapEndpoints(service {{.PkgName}}.{{.ServiceName}}Server, opt *grpc_tool.Options) *Endpoints {
	serviceName := {{.PkgName}}.ServiceName
	return &Endpoints{
{{- range .Methods}}
	{{.Name}}Endpoint: wrapEndpoint(serviceName, "{{.Name}}",wrap{{.Name}}(service), opt),
{{- end}}
	}
}`
	transportTemplate = `
package transport

import (
	"context"

	"{{.ImportPrefix}}/grpc/endpoints"
	"sobe-kit/grpc_tool"

	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/transport/grpc"

	{{.PkgName}} "{{.ImportPath}}"
)

type gRPCServer struct {
{{- range .Methods}}
	{{FirstLower .Name}}Handler grpc.Handler
{{- end}}
}

{{range .Methods}}
func (g *gRPCServer) {{.Name}}(ctx context.Context, req *{{$.PkgName}}.{{.RequestName}}) (*{{$.PkgName}}.{{.ResponseName}}, error) {
	_, ret, err := g.{{FirstLower .Name}}Handler.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return ret.(*{{$.PkgName}}.{{.ResponseName}}), nil
}
{{end}}

func NewGRPCServer(service {{.PkgName}}.{{.ServiceName}}Server, opts ...grpc_tool.Option) {{.PkgName}}.{{.ServiceName}}Server {
	options := grpc_tool.NewOptions()
	for _, opt := range opts {
		opt(options)
	}

	serviceName := {{.PkgName}}.ServiceName
	eps := endpoints.WrapEndpoints(service,options)
	return &gRPCServer{
	{{- range .Methods}}
		{{FirstLower .Name}}Handler: grpc.NewServer(
			eps.{{.Name}}Endpoint,
			DecodeRequestFunc,
			EncodeResponseFunc,
			append(grpc_tool.GetServerOptions(serviceName, "{{.Name}}"), grpc.ServerBefore(opentracing.GRPCToContext(options.Trace, "{{.Name}}", options.Logger)))...),
	{{end}}
	}
}

func DecodeRequestFunc(_ context.Context, req interface{}) (request interface{}, err error) {
	request = req
	return
}

func EncodeResponseFunc(_ context.Context, resp interface{}) (response interface{}, err error) {
	response = resp
	return
}
`
	clientTemplate = `
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"

    "sobe-kit/grpc_tool"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
    "github.com/go-kit/kit/tracing/opentracing"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	trans "github.com/go-kit/kit/transport/grpc"
	edp "{{.ImportPrefix}}/grpc/endpoints"
	{{.PkgName}} "{{.ImportPath}}"
)

func NewClient(opts ...grpc_tool.Option) ({{.PkgName}}.{{.ServiceName}}Server, error) {
	o := grpc_tool.NewOptions()
	for _, opt := range opts {
		opt(o)
	}

	grpcOpts, err := gRPCDialOptions(o.ServerName, o.Cert)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial("{{.PkgName}}:{{.Port}}", grpcOpts...)
	if err != nil {
		return nil, err
	}

	wrap := wrapMethod({{.PkgName}}.ServiceName, conn, o)
	return &edp.Endpoints{
	{{- range .Methods}}
		{{.Name}}Endpoint: wrap("{{.Name}}", new({{$.PkgName}}.{{.ResponseName}})),
	{{- end}}
	}, nil
}

func wrapMethod(serviceName string, conn *grpc.ClientConn, opt *grpc_tool.Options) func(method string, reply interface{}) endpoint.Endpoint {
	return func(method string, reply interface{}) endpoint.Endpoint {
		ep := trans.NewClient(
			conn,
			serviceName,
			method,
			encodeGRPCRequest,
			decodeGRPCResponse,
			reply,
			append(grpc_tool.GetClientOptions(serviceName, method), trans.ClientBefore(opentracing.ContextToGRPC(opt.Trace, opt.Logger)))...).Endpoint()

		return wrapEndpoint(
			serviceName,
			method,
			ep)
	}
}

// consul load balance
func NewClientWithConsul(consulAddr, dataCenter string, opts ...grpc_tool.Option) ({{.PkgName}}.{{.ServiceName}}Server, error) {
	o := grpc_tool.NewOptions()
	for _, opt := range opts {
		opt(o)
	}

	consulClient, err := api.NewClient(&api.Config{
		Address:    consulAddr,
		Datacenter: dataCenter,
	})
	if err != nil {
		return nil, err
	}

	grpcOpts, err := gRPCDialOptions(o.ServerName, o.Cert)
	if err != nil {
		return nil, err
	}

	instanter := consul.NewInstancer(consul.NewClient(consulClient), o.Logger, {{.PkgName}}.ServiceName, o.Tags, true)
	retryWrap := retryEndpoint(instanter, {{.PkgName}}.ServiceName, o.Logger, grpcOpts...)
	return &edp.Endpoints{
	{{- range .Methods}}
		{{.Name}}Endpoint: retryWrap("{{.Name}}", new({{$.PkgName}}.{{.ResponseName}})),
	{{- end}}
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
}`

	healthTemplate = `package service

import (
	"context"

	{{.PkgName}} "{{.ImportPath}}"
)

func (s *service) Health(_ context.Context, _ *{{.PkgName}}.HealthRequest) (*{{.PkgName}}.HealthResponse, error) {
	return &{{.PkgName}}.HealthResponse{}, nil
}`

	serviceTemplate = `package service

import (
	"io"

    "sobe-kit/props"

	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	{{.PkgName}} "{{.ImportPath}}"
)

type {{FirstUpper .ServiceName}}Service interface {
	io.Closer
	{{.PkgName}}.{{FirstUpper .ServiceName}}Server
    grpc_health_v1.HealthServer
}

type Config struct {
    Port int    {{.Quote}}json:"port"{{.Quote}} 
	Name string {{.Quote}}json:"name"{{.Quote}}
	Age  int    {{.Quote}}json:"age"{{.Quote}}
}

type service struct {
	*health.Server
}

func New{{FirstUpper .ServiceName}}(trace opentracing.Tracer, log log.Logger) ({{FirstUpper .ServiceName}}Service, error) {
    var cfg Config
	err := props.ConfigFromFile(props.JsonDecoder(&cfg), props.Always, props.FilePath("./conf/{{.PkgName}}.conf"))
	if err != nil {
		return nil, err
	}

	checkServer := health.NewServer()
	checkServer.SetServingStatus("health", grpc_health_v1.HealthCheckResponse_SERVING)
	return &service{
		Server: checkServer,
	}, nil
}

func (s *service) Close() error {
	return nil
}
`

	serverTemplate = `package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	"sobe-kit/deviceinfo"
	"sobe-kit/grpc_tool"
	"sobe-kit/props"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"{{.ImportPath}}/transport"	
	"{{ServicePath .ImportPath}}"

	{{.PkgName}} "{{.ImportPath}}"
)

type Config struct {
	Address     string                {{.Quote}}json:"address"{{.Quote}}     // 监听地址
	Certificate *grpc_tool.Cert       {{.Quote}}json:"certificate"{{.Quote}} // 证书信息
	Consul      *grpc_tool.ConsulConf {{.Quote}}json:"consul"{{.Quote}}     // consul注册中心
}

func Server() {
	glog.CopyStandardLogTo("ERROR")
	rand.Seed(time.Now().UnixNano())
	defer glog.Flush()

	var cfg Config
	err := props.LoadConfig(&cfg)
	if err != nil {
		glog.Fatal(err)
	}

	cfg, err = DefaultConfig(cfg)
	if err != nil {
		glog.Fatal(err)
	}

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		glog.Fatal(err)
	}

	var opts []grpc.ServerOption
	if cfg.Certificate != nil {
		cert, err := grpc_tool.LoadCertificates(cfg.Certificate)
		if err != nil {
			glog.Fatal(err)
		}
		opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(cert)))
	}

	grpcServer := grpc.NewServer(opts...)
	svr, err := service.New{{FirstUpper .ServiceName}}()
    if err != nil {
		log.Fatal(err)
	}
	{{.PkgName}}.Register{{FirstUpper .ServiceName}}Server(grpcServer, transport.NewGRPCServer(svr))

	if cfg.Consul != nil {
		deregister, err := grpc_tool.RegisterServiceWithConsul(cfg.Consul)
		if err != nil {
			glog.Fatal(err)
		}

		defer func() {
			err = deregister()
			if err != nil {
				glog.Errorf("deregister Err:%v", err)
			}
		}()
	}

	quit := graceQuit(func() {
		grpcServer.GracefulStop()
	})

	log.Printf("%s started with http at %v\n", "{{.PkgName}}", cfg.Address)
	err = grpcServer.Serve(listener)
	if err != nil {
		glog.Fatal(err)
		return
	}
	<-quit
}

func graceQuit(do func()) <-chan struct{} {
	quit := make(chan struct{})
	go func() {
		ch := make(chan os.Signal, 10)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		select {
		case <-ch:
			signal.Stop(ch)
			do()
			quit <- struct{}{}
		}
	}()
	return quit
}

func DefaultConfig(cfg Config) (Config, error) {
	var err error
	if cfg.Consul != nil {
		if cfg.Consul.Registration.Name == "" {
			cfg.Consul.Registration.Name = {{.PkgName}}.ServiceName
		}

		if cfg.Consul.Registration.ID == "" {
			cfg.Consul.Registration.Name = fmt.Sprintf("%v-%v-%v", deviceinfo.GetAppName(), deviceinfo.GetLANHost(), cfg.Address)
		}

		cfg.Consul.Registration.Port, err = deviceinfo.GetPort(cfg.Address)
		if err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

func CATls(rootCa, serverCa, serverKey string) grpc.ServerOption {
	cert, err := tls.LoadX509KeyPair(serverCa, serverKey)
	if err != nil {
		log.Fatal(err)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(rootCa)
	if err != nil {
		log.Fatal(err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal(err)
	}

	return grpc.Creds(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}))
}
`
	serverClientTemplate = `
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"

	"github.com/golang/glog"
	"sobe-kit/grpc_tool"
	"{{ServerPath .ImportPath}}"
	"sobe-kit/props"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	{{.PkgName}} "{{.ImportPath}}"
)

func Client() {
	var cfg server.Config
	err := props.LoadConfig(&cfg)
	if err != nil {
		glog.Fatal(err)
	}

	cfg, err = server.DefaultConfig(cfg)
	if err != nil {
		glog.Fatal(err)
	}

	var opts []grpc.DialOption
	if cfg.Certificate != nil {
		cert, err := grpc_tool.LoadCertificates(cfg.Certificate)
		if err != nil {
			glog.Fatal(err)
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewServerTLSFromCert(cert)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(cfg.Address, opts...)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	cli := {{.PkgName}}.New{{FirstUpper .PkgName}}Client(conn)

	// ping
	ret, err := cli.Ping(context.Background(), &{{.PkgName}}.PingRequest{
		Service: "{{.ServiceName}}",
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("Ping Response:%#v\n", *ret)
}

func CATls(rootCa, clientCa, clientKey, hostName string) grpc.DialOption {
	cert, err := tls.LoadX509KeyPair(clientCa, clientKey)
	if err != nil {
		log.Fatal(err)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(rootCa)
	if err != nil {
		log.Fatal(err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal(err)
	}

	c := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   "localhost",
		RootCAs:      certPool,
	})
	return grpc.WithTransportCredentials(c)
}
`

	cmdTemplate = `
	package cmd

import (
	"log"

	"sobe-kit/props"
	"{{ServerClientPath .ImportPath}}"
	"{{ServerPath .ImportPath}}"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "{{.ServiceName}} service cmd",
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "start {{.ServiceName}} client",
	Run: func(cmd *cobra.Command, args []string) {
		client.Client()
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start {{.ServiceName}} server",
	Run: func(cmd *cobra.Command, args []string) {
		server.Server()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	// root
	rootCmd.PersistentFlags().StringVarP(&props.ConfFilePath, "config-path", "", "./conf/{{.ServiceName}}-local.conf", "config file path")
	rootCmd.PersistentFlags().StringVarP(&props.ConsulAddress, "consul-addr", "", "", "consul address")
	rootCmd.PersistentFlags().StringVarP(&props.ConsulSchema, "consul-schema", "", "", "consul schema")
	rootCmd.PersistentFlags().StringVarP(&props.ConsulDataCenter, "consul-data-center", "", "", "consul data center")
	rootCmd.PersistentFlags().Int64VarP(&props.ConsulWaitTime, "consul-wait-time", "", 0, "consul wait time")
	rootCmd.PersistentFlags().StringVarP(&props.ConsulToken, "consul-token", "", "", "consul token")
	rootCmd.PersistentFlags().StringVarP(&props.ConfNode, "config-node", "", "", "config node in consul")
	// server and client
	rootCmd.AddCommand(serverCmd, clientCmd)
}
`

	mainTemplate = `
package main

import (
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"sobe-kit/grpc_tool"
	"{{.ImportPrefix}}/service"
	"{{.ImportPrefix}}/grpc/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	{{.PkgName}} "{{.ImportPath}}"
    logger "github.com/go-kit/kit/log"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	listener, err := net.Listen("tcp", os.Getenv("SERVER_ADDRESS"))
	if err != nil {
		log.Fatal(err)
	}

	var opts []grpc_tool.Option
	if metricAddress := os.Getenv("METRIC_ADDRESS"); len(metricAddress) > 0 {
		grpc_tool.NewExposer(metricAddress)
		opts = append(opts, grpc_tool.WithCollect(grpc_tool.NewMetricsCollector()))
	}

	trace, traceCloser := grpc_tool.NewJaegerTracer("{{.ServiceName}}", os.Getenv("TRACE_ADDRESS"))
	l := logger.NewNopLogger()

	handle, err := service.New{{.ServiceName}}(trace, l)
	if err != nil {
		log.Fatal(err)
	}

	grpcSvr := grpc.NewServer()
	{{.PkgName}}.Register{{.ServiceName}}Server(grpcSvr,transport.NewGRPCServer(handle, append(opts, grpc_tool.WithTrace(trace), grpc_tool.WithLogger(l))...))
	grpc_health_v1.RegisterHealthServer(grpcSvr, handle)

	grpc_tool.Graceful(func() {
		grpcSvr.GracefulStop()
		err := handle.Close()
		if err != nil {
			log.Println(err)
		}

		err = traceCloser.Close()
		if err != nil {
			log.Println(err)
		}
	})

	log.Printf("%v started with %v", "order", listener.Addr().String())
	err = grpcSvr.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}	
}
`

	configTemplate = `{
	"port": {{.Port}},
	"name": "{{.PkgName}}",
	"age": 18
}`

	buildTemplate = `#!/bin/sh
IMAGE_TAG=v0.0.1
SERVICE_NAME={{.PkgName}}
IMAGE_NAME={{.Registry}}${SERVICE_NAME}

set -o errexit
GOOS=linux GOARCH=amd64 go build -i -o _bin/${SERVICE_NAME}
docker build . -t ${IMAGE_NAME}:${IMAGE_TAG}
docker push ${IMAGE_NAME}:${IMAGE_TAG}
rm -f _bin/${SERVICE_NAME}`

	dockerFileTemplate = `FROM alpine:3.12.0
RUN apk add --update ca-certificates && \
    rm -rf /var/cache/apk/* /tmp/*
ENV PORT {{.Port}}
ENV SERVER_NAME {{.ServiceName}}
ENV SERVER_ADDRESS 0.0.0.0:{{.Port}}
{{if gt .MetricPort 0}}ENV METRIC_ADDRESS 0.0.0.0:{{.MetricPort}}{{end}}
WORKDIR /app
COPY conf/{{.PkgName}}.conf /app/conf/{{.PkgName}}.conf
ADD _bin/{{.PkgName}} /app
EXPOSE {{.Port}}
{{if gt .MetricPort 0}}EXPOSE {{.MetricPort}}{{end}}
CMD [ "./{{.PkgName}}" ]
`

	makefileTemplate = `PROJECT_NAME=$(notdir $(shell pwd))

vpath $(PROJECT_NAME) ./_bin
vpath %.proto ./grpc
vpath %.pb.go ./grpc/pb

PROTO_PATH=./grpc
PB_TARGET=./grpc

export SERVER_ADDRESS=0.0.0.0:{{.Port}}
{{if gt .MetricPort 0}}export METRIC_ADDRESS=0.0.0.0:{{.MetricPort}}{{end}}

$(PROJECT_NAME).pb.go: $(PROJECT_NAME).proto
	@echo "building *.pb.go"
	protoc --proto_path=$(PROTO_PATH) --go_out=plugins=grpc:$(PB_TARGET) $(PROTO_PATH)/*.proto
	sobe-kit -g -p=./../{{.PkgName}}

$(PROJECT_NAME):  $(PROJECT_NAME).pb.go
	@echo "build " $(PROJECT_NAME)
	if [ ! -e _bin ]; then mkdir _bin; fi;
	go build -i -o _bin/$(PROJECT_NAME)

.PHONY: gen 
gen:
	@echo "building *.pb.go"
	protoc --proto_path=$(PROTO_PATH) --go_out=plugins=grpc:$(PB_TARGET) $(PROTO_PATH)/*.proto
	sobe-kit -g -p=./../{{.PkgName}}

.PHONY: check
check:
	go vet ./...
	go test -race $$(go list ./...| grep -v pkg)

.PHONY: build
build: $(PROJECT_NAME)
	@echo successful

.PHONY: run
run: $(PROJECT_NAME)
	@echo "start "  $(PROJECT_NAME)
	for i in $$(ps -ef | grep _bin/$(PROJECT_NAME) | grep -v grep | awk '{print $$2}') ; do \
	   kill $$i;\
	done
	nohup _bin/$(PROJECT_NAME) >> _bin/nohup.out 2>&1 &

.PHONY: clean
clean:
	rm -rf _bin
	rm -rf ./api
	rm -rf ./grpc/pb
	rm -rf ./grpc/client
	rm -rf ./grpc/endpoints
	rm -rf ./grpc/transport

.PHONY: check
check:
	make gen 
	rm -rf _bin
	sh ./build.sh`

	deploymentTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: {{.Namespace}} 
  name: {{.PkgName}}
spec:
  replicas: 1
  selector:
    matchLabels:
      name: {{.PkgName}}
  template:
    metadata:
      labels:
        name: {{.PkgName}} 
    spec:
      containers:
        - name: {{.PkgName}} 
          check: {{.Registry}}{{.PkgName}}:v0.0.1
          imagePullPolicy: Always
          ports:
			- name: {{.PkgName}}
              containerPort: {{.Port}}
          volumeMounts:
            - name: {{.PkgName}}.configmap 
              mountPath: /app/conf
        - name: health
          check: {{.Registry}}health:v0.0.1
          imagePullPolicy: Always
          command: [
            "/health",
            "--health_address=0.0.0.0:8080",
            "--server_address=127.0.0.1:{{.Port}}"
          ]
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 3
            periodSeconds: 3
      volumes:
        - name: {{.PkgName}}.configmap
          configMap:
            name: {{.PkgName}}
`
	configMapTemplate = `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.PkgName}}
  namespace: {{.Namespace}} 
data:
  {{.PkgName}}.conf: |
    {
      "name": "{{.PkgName}}",
      "age": 18,
      "port": {{.Port}}
    }
`

	k8sServiceTemplate = `apiVersion: v1
kind: Service
metadata:
  name: {{.PkgName}}
  namespace: {{.Namespace}} 
spec:
  selector:
    name: {{.PkgName}}
  ports:
    - port: {{.Port}} 
      protocol: TCP
`
)
