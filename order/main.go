package main

import (
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"sobe-kit/grpc_tool"
	"sobe-kit/order/grpc/transport"
	"sobe-kit/order/service"

	order "sobe-kit/order/grpc/pb"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	handle, err := service.NewOrder()
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", os.Getenv("SERVER_ADDRESS"))
	if err != nil {
		log.Fatal(err)
	}

	grpcSvr := grpc.NewServer()
	order.RegisterOrderServer(grpcSvr, transport.NewGRPCServer(handle))

	grpc_tool.Graceful(func() {
		grpcSvr.GracefulStop()
		err := handle.Close()
		if err != nil {
			log.Fatal(err)
		}
	})

	log.Printf("%v started with %v", "order", listener.Addr().String())
	err = grpcSvr.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}
