package server

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
	"sobe-kit/example/timi/grpc/transport"
	"sobe-kit/example/timi/service"
	"sobe-kit/grpc_tool"
	"sobe-kit/props"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	timi "sobe-kit/example/timi/grpc"
)

type Config struct {
	Address     string                `json:"address"`     // 监听地址
	Certificate *grpc_tool.Cert       `json:"certificate"` // 证书信息
	Consul      *grpc_tool.ConsulConf `json:"consul"`      // consul注册中心
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
	svr := service.NewTimi()
	timi.RegisterTimiServer(grpcServer, transport.NewGRPCServer(svr))

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

	log.Printf("%s started with http at %v\n", "timi", cfg.Address)
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
			cfg.Consul.Registration.Name = timi.ServiceName
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
