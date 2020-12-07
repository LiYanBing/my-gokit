package service

import (
	"context"

	order "sobe-kit/order/grpc/pb"
)

func (s *service) Health(_ context.Context, _ *order.HealthRequest) (*order.HealthResponse, error) {
	return &order.HealthResponse{}, nil
}
