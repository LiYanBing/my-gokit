package service

import (
	"io"

	"sobe-kit/props"

	order "sobe-kit/order/grpc/pb"
)

type OrderService interface {
	io.Closer
	order.OrderServer
}

type Config struct {
	Port int    `json:"port"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type service struct {
}

func NewOrder() (OrderService, error) {
	var cfg Config
	err := props.ConfigFromFile(props.JsonDecoder(&cfg), props.Always, props.FilePath("./conf/order.conf"))
	if err != nil {
		return nil, err
	}
	return &service{}, nil
}

func (s *service) Close() error {
	return nil
}
