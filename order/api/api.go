package order

import (
	"context"

	order "sobe-kit/order/grpc/pb"
)

type Order interface {
	Health(context.Context, *order.HealthRequest) (*order.HealthResponse, error)
}
