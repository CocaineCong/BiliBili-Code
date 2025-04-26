package tcc

import (
	"context"
	"fmt"
	"sync"
)

// Participant TCC事务参与者接口
type Participant interface {
	Try(ctx context.Context) error
	Confirm(ctx context.Context) error
	Cancel(ctx context.Context) error
}

// Coordinator TCC事务协调者
type Coordinator struct {
	participants []Participant
}

func NewCoordinator(participants ...Participant) *Coordinator {
	return &Coordinator{
		participants: participants,
	}
}

// Execute 执行TCC事务
func (c *Coordinator) Execute(ctx context.Context) error {
	// 阶段1: Try
	var tryErr error
	for _, p := range c.participants {
		if err := p.Try(ctx); err != nil {
			// Try阶段失败，执行Cancel
			tryErr = err
		}
	}
	if tryErr != nil {
		c.cancelAll(ctx)
		return tryErr
	}

	// 阶段2: Confirm
	var confirmErr error
	for _, p := range c.participants {
		if err := p.Confirm(ctx); err != nil {
			confirmErr = fmt.Errorf("confirm phase failed: %w", err)
			break
		}
	}

	if confirmErr != nil {
		// Confirm阶段失败，执行Cancel
		c.cancelAll(ctx)
		return fmt.Errorf("confirm phase failed, rolled back: %w", confirmErr)
	}

	return nil
}

// cancelAll 执行所有参与者的Cancel操作
func (c *Coordinator) cancelAll(ctx context.Context) {
	var wg sync.WaitGroup
	errChan := make(chan error, len(c.participants))

	for _, p := range c.participants {
		wg.Add(1)
		go func(p Participant) {
			defer wg.Done()
			if err := p.Cancel(ctx); err != nil {
				errChan <- fmt.Errorf("cancel failed: %w", err)
			}
		}(p)
	}

	// 等待所有Cancel操作完成
	wg.Wait()
	close(errChan)

	// 收集所有Cancel错误
	var cancelErrs []error
	for err := range errChan {
		cancelErrs = append(cancelErrs, err)
	}

	// 如果有Cancel失败，记录错误
	if len(cancelErrs) > 0 {
		fmt.Printf("some cancellations failed: %v\n", cancelErrs)
	}
}
