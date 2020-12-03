package gokit_tool

const (
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

func wrapEndpoint(service, method string, endpoint endpoint.Endpoint) endpoint.Endpoint {
	wrappers := grpc_tool.GetEndpointMiddleware(service, method)
	for _, wrap := range wrappers {
		endpoint = wrap(endpoint)
	}
	return endpoint
}

func WrapEndpoints(serviceName string, service {{.PkgName}}.{{.ServiceName}}Server) *Endpoints {
	return &Endpoints{
{{- range .Methods}}
	{{.Name}}Endpoint: wrapEndpoint(serviceName, "{{.Name}}", wrap{{.Name}}(service)),
{{- end}}
	}
}`
	transportTemplate = `
package transport

import (
	"context"

	"github.com/go-kit/kit/transport/grpc"
	"{{.ImportPath}}/endpoints"
	"sobe-kit/grpc_tool"

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

func NewGRPCServer(service {{.PkgName}}.{{.ServiceName}}Server) {{.PkgName}}.{{.ServiceName}}Server {
	eps := endpoints.WrapEndpoints({{.PkgName}}.ServiceName, service)
	return &gRPCServer{
	{{- range .Methods}}
		{{FirstLower .Name}}Handler: grpc.NewServer(
			eps.{{.Name}}Endpoint,
			DecodeRequestFunc,
			EncodeResponseFunc,
			grpc_tool.GetServerOptions({{$.PkgName}}.ServiceName, "{{.Name}}")...),
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

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	"github.com/hashicorp/consul/api"
	"sobe-kit/grpc_tool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	trans "github.com/go-kit/kit/transport/grpc"
	edp "{{.ImportPath}}/endpoints"
	{{.PkgName}} "{{.ImportPath}}"
)

func NewClient(addr, serverName string, cert []byte) ({{.PkgName}}.{{.ServiceName}}Server, error) {
	opts, err := gRPCDialOptions(serverName, cert)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	wrap := wrapMethod(serverName, conn)
	return &edp.Endpoints{
	{{- range .Methods}}
		{{.Name}}Endpoint: wrap("{{.Name}}", new({{$.PkgName}}.{{.ResponseName}})),
	{{- end}}
	}, nil
}

func wrapMethod(serviceName string, conn *grpc.ClientConn) func(method string, reply interface{}) endpoint.Endpoint {
	return func(method string, reply interface{}) endpoint.Endpoint {
		ep := trans.NewClient(
			conn,
			serviceName,
			method,
			encodeGRPCRequest,
			decodeGRPCResponse,
			reply,
			grpc_tool.GetClientOptions(serviceName, method)...).Endpoint()

		return wrapEndpoint(
			serviceName,
			method,
			ep)
	}
}

// consul load balance
func NewClientWithConsul(consulAddr, dataCenter, serverName string, cert []byte, tags []string, logger log.Logger) ({{.PkgName}}.{{.ServiceName}}Server, error) {
	consulClient, err := api.NewClient(&api.Config{
		Address:    consulAddr,
		Datacenter: dataCenter,
	})
	if err != nil {
		return nil, err
	}

	opts, err := gRPCDialOptions(serverName, cert)
	if err != nil {
		return nil, err
	}

	instanter := consul.NewInstancer(consul.NewClient(consulClient), logger, {{.PkgName}}.ServiceName, tags, true)
	retryWrap := retryEndpoint(instanter, serverName, logger, opts...)
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

	serviceTemplate = `package service

import (
	"context"

	{{.PkgName}} "{{.ImportPath}}"
)

type service struct {
}

func New{{FirstUpper .ServiceName}}() {{.PkgName}}.{{.ServiceName}}Server {
	return &service{}
}

func (s *service) HelloWorld(ctx context.Context, req *{{.PkgName}}.HelloWorldRequest) (*{{.PkgName}}.HelloWorldResponse, error) {
	return &{{.PkgName}}.HelloWorldResponse{
		Output: "Hello " + req.Input,
	}, nil
}

func (s *service) Ping(ctx context.Context, req *{{.PkgName}}.PingRequest) (*{{.PkgName}}.PingResponse, error) {
	return &{{.PkgName}}.PingResponse{
		Status: "pong " + req.Service,
	}, nil
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
	svr := service.New{{FirstUpper .ServiceName}}()
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

	// hello world
	r, err := cli.HelloWorld(context.Background(), &{{.PkgName}}.HelloWorldRequest{
		Input: "{{.ServiceName}}",
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("HelloWorld Response:%#v\n", *r)

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
	"{{CmdPath .ImportPath}}"
)

func main() {
	cmd.Execute()
}
`
)
