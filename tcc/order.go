package tcc

import (
	"context"
)

// OrderService 订单服务
type OrderService struct {
	orderID     string
	status      string
	confirmFunc func(ctx context.Context) error
}

func NewOrderService(orderID string) *OrderService {
	return &OrderService{
		orderID: orderID,
		status:  "pending",
	}
}

func (o *OrderService) Try(ctx context.Context) error {
	// 尝试创建订单
	o.status = "trying"
	return nil
}

func (o *OrderService) Confirm(ctx context.Context) error {
	// 确认订单
	if o.confirmFunc != nil {
		return o.confirmFunc(ctx)
	}
	o.status = "confirmed"
	return nil
}

func (o *OrderService) Cancel(ctx context.Context) error {
	// 取消订单
	o.status = "canceled"
	return nil
}
