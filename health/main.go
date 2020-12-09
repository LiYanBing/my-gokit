package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

var serverAddress string
var healthAddress string

func init() {
	flag.StringVar(&serverAddress, "server-address", "127.0.0.1:2048", "server address of grpc health")
	flag.StringVar(&healthAddress, "health-address", "0.0.0.0:8080", "health address of http")
}

func main() {
	flag.Parse()
	grpcConn, err := GetGRPCConn(serverAddress)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if grpcConn != nil {
			grpcConn.Close()
		}
	}()

	err = http.ListenAndServe(healthAddress, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		var reply grpc_health_v1.HealthCheckResponse
		err = grpcConn.Invoke(context.Background(), "/grpc.health.v1.Health/Check", &grpc_health_v1.HealthCheckRequest{
			Service: "health",
		}, &reply)
		if err != nil {
			fmt.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("ok"))
	}))
	if err != nil {
		log.Fatal(err)
	}
}

// GRPC connection
func GetGRPCConn(addr string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 10,
			Timeout:             time.Second * 5,
			PermitWithoutStream: false,
		}),
	}
	return grpc.Dial(addr, opts...)
}
