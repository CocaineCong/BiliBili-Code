package tcc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaymentTCCSuccess(t *testing.T) {
	// 准备测试数据
	account := NewAccountService("FanOne", 1000.0)
	inventory := NewInventoryService("iPhone18", 10)
	// 准备业务数据
	bizData := struct {
		OrderID  string
		Amount   float64
		Quantity int
	}{
		OrderID:  "FanOne-Apple-Success",
		Amount:   500.0,
		Quantity: 2,
	}

	// 设置预留资源
	account.PrepareTry(bizData.Amount)
	inventory.PrepareTry(bizData.Quantity)
	// 创建TCC协调者
	coordinator := NewCoordinator(account, inventory)
	// 执行TCC事务
	err := coordinator.Execute(context.Background())
	assert.NoError(t, err)
	// 验证结果
	assert.Equal(t, 500.0, account.amount)
	assert.Equal(t, 0.0, account.frozen)
	assert.Equal(t, 8, inventory.quantity)
	assert.Equal(t, 0, inventory.frozen)
}

func TestPaymentTCCFailInsufficientBalance(t *testing.T) {
	// 准备测试数据
	account := NewAccountService("FanOne", 100.0)
	inventory := NewInventoryService("MacBookPro", 10)

	// 准备业务数据
	bizData := struct {
		OrderID  string
		Amount   float64
		Quantity int
	}{
		OrderID:  "FanOne-MAC-Failed",
		Amount:   500.0,
		Quantity: 2,
	}
	// 设置预留资源
	account.PrepareTry(bizData.Amount)
	inventory.PrepareTry(bizData.Quantity)
	// 创建TCC协调者
	coordinator := NewCoordinator(account, inventory)
	// 执行TCC事务
	err := coordinator.Execute(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")
	// 验证结果
	assert.Equal(t, 100.0, account.amount)
	assert.Equal(t, 0.0, account.frozen)
	assert.Equal(t, 10, inventory.quantity)
	assert.Equal(t, 0, inventory.frozen)
}

func TestPaymentTCCConfirmFailure(t *testing.T) {
	// 准备测试数据
	account := NewAccountService("FanOne", 1000.0)
	inventory := NewInventoryService("MacBookPro", 10)
	order := NewOrderService("FanOne-MAC-ConfirmFailed")
	order.confirmFunc = func(ctx context.Context) error {
		return errors.New("order confirm failed")
	}
	// 准备业务数据
	bizData := struct {
		Amount   float64
		Quantity int
	}{
		Amount:   500.0,
		Quantity: 2,
	}
	// 设置预留资源
	account.PrepareTry(bizData.Amount)
	inventory.PrepareTry(bizData.Quantity)
	// 创建TCC协调者
	coordinator := NewCoordinator(account, inventory, order)
	// 执行TCC事务
	err := coordinator.Execute(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order confirm failed")
	// 验证状态
	assert.Equal(t, 1000.0, account.amount) // 应该回滚到初始状态
	assert.Equal(t, 0.0, account.frozen)
	assert.Equal(t, 10, inventory.quantity) // 应该回滚到初始状态
	assert.Equal(t, 0, inventory.frozen)
	assert.Equal(t, "canceled", order.status) // 订单状态应该保持pending
}
