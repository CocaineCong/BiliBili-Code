package tcc

import (
	"context"
	"errors"
)

// AccountService 账户服务
type AccountService struct {
	accountID    string
	amount       float64
	frozen       float64
	deductAmount float64 // 添加要扣减的金额字段
}

func NewAccountService(accountID string, balance float64) *AccountService {
	return &AccountService{
		accountID: accountID,
		amount:    balance,
	}
}

// PrepareTry 预留资源
func (a *AccountService) PrepareTry(amount float64) {
	a.deductAmount = amount
}

// Try 实现 TCC 的 Try 阶段
func (a *AccountService) Try(ctx context.Context) error {
	a.frozen = a.deductAmount
	a.amount -= a.deductAmount
	if a.amount < 0 {
		return errors.New("insufficient balance")
	}
	return nil
}

func (a *AccountService) Confirm(ctx context.Context) error {
	// 确认操作，实际业务中可能将冻结金额转出
	a.frozen = 0
	return nil
}

func (a *AccountService) Cancel(ctx context.Context) error {
	// 取消操作，返还金额
	a.amount += a.deductAmount
	a.frozen = 0
	return nil
}
