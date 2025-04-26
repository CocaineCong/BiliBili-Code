package tcc

import (
	"context"
	"errors"
)

// InventoryService 库存服务
type InventoryService struct {
	productID      string
	quantity       int
	frozen         int
	deductQuantity int // 添加要扣减的数量字段
}

func NewInventoryService(productID string, quantity int) *InventoryService {
	return &InventoryService{
		productID: productID,
		quantity:  quantity,
	}
}

// PrepareTry 预留资源
func (i *InventoryService) PrepareTry(quantity int) {
	i.deductQuantity = quantity
}

// Try 实现 TCC 的 Try 阶段
func (i *InventoryService) Try(ctx context.Context) error {
	i.frozen = i.deductQuantity
	i.quantity -= i.deductQuantity
	if i.quantity < 0 {
		return errors.New("insufficient inventory")
	}
	return nil
}

func (i *InventoryService) Confirm(ctx context.Context) error {
	// 确认操作，实际业务中可能需要更新数据库等操作
	i.frozen = 0
	return nil
}

func (i *InventoryService) Cancel(ctx context.Context) error {
	// 取消操作，返还库存
	i.quantity += i.deductQuantity
	i.frozen = 0
	return nil
}
