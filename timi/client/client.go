package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"

	"github.com/golang/glog"
	"github.com/liyanbing/my-gokit/props"
	"github.com/liyanbing/my-gokit/timi/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	timi "github.com/liyanbing/my-gokit/timi/grpc"
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
	//if cfg.Certificate != nil {
	//	cert, err := grpc_tool.LoadCertificates(cfg.Certificate)
	//	if err != nil {
	//		glog.Fatal(err)
	//	}
	//	opts = append(opts, grpc.WithTransportCredentials(credentials.NewServerTLSFromCert(cert)))
	//} else {
	//	opts = append(opts, grpc.WithInsecure())
	//}
	//cret, err := credentials.NewClientTLSFromFile("./certs/server.pem", "localhost")
	//if err != nil {
	//	glog.Fatal(err)
	//}

	cert, err := tls.LoadX509KeyPair("./certs/client/client.pem", "./certs/client/client.key")
	if err != nil {
		log.Fatal(err)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("./certs/ca.pem")
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
	opts = append(opts, grpc.WithTransportCredentials(c))

	conn, err := grpc.Dial(cfg.Address, opts...)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	cli := timi.NewTimiClient(conn)

	// hello world
	r, err := cli.HelloWorld(context.Background(), &timi.HelloWorldRequest{
		Input: "Timi",
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("HelloWorld Response:%#v\n", *r)

	// ping
	ret, err := cli.Ping(context.Background(), &timi.PingRequest{
		Service: "Timi",
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("Ping Response:%#v\n", *ret)
}
