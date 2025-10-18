package limit

import (
	"sync"
	"testing"
	"time"
)

// TestFixedWindowCounter_Basic 测试固定窗口计数器基本功能
func TestFixedWindowCounter_Basic(t *testing.T) {
	limiter := NewFixedWindowCounter(3, time.Second)
	
	// 前3个请求应该通过
	for i := 0; i < 3; i++ {
		if !limiter.Allow() {
			t.Errorf("第%d个请求应该通过", i+1)
		}
	}
	
	// 第4个请求应该被拒绝
	if limiter.Allow() {
		t.Error("第4个请求应该被拒绝")
	}
	
	// 检查状态
	current, limit := limiter.GetStatus()
	if current != 3 || limit != 3 {
		t.Errorf("状态错误: current=%d, limit=%d", current, limit)
	}
}

// TestFixedWindowCounter_WindowReset 测试窗口重置功能
func TestFixedWindowCounter_WindowReset(t *testing.T) {
	limiter := NewFixedWindowCounter(2, 100*time.Millisecond)
	
	// 消耗所有令牌
	limiter.Allow()
	limiter.Allow()
	
	// 应该被拒绝
	if limiter.Allow() {
		t.Error("请求应该被拒绝")
	}
	
	// 等待窗口重置
	time.Sleep(150 * time.Millisecond)
	
	// 现在应该可以通过
	if !limiter.Allow() {
		t.Error("窗口重置后请求应该通过")
	}
	
	// 检查状态，应该重置为1
	current, _ := limiter.GetStatus()
	if current != 1 {
		t.Errorf("窗口重置后计数应该为1，实际为%d", current)
	}
}

// TestFixedWindowCounter_Concurrent 测试并发安全性
func TestFixedWindowCounter_Concurrent(t *testing.T) {
	limiter := NewFixedWindowCounter(100, time.Second)
	var wg sync.WaitGroup
	var successCount int64
	var mu sync.Mutex
	
	// 启动200个并发请求
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.Allow() {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}
	
	wg.Wait()
	
	// 应该只有100个请求成功
	if successCount != 100 {
		t.Errorf("期望100个成功请求，实际%d个", successCount)
	}
	
	// 验证状态
	current, limit := limiter.GetStatus()
	if current != 100 || limit != 100 {
		t.Errorf("最终状态错误: current=%d, limit=%d", current, limit)
	}
}

// TestFixedWindowCounter_ZeroLimit 测试零限制
func TestFixedWindowCounter_ZeroLimit(t *testing.T) {
	limiter := NewFixedWindowCounter(0, time.Second)
	
	// 任何请求都应该被拒绝
	if limiter.Allow() {
		t.Error("零限制时请求应该被拒绝")
	}
	
	current, limit := limiter.GetStatus()
	if current != 0 || limit != 0 {
		t.Errorf("零限制状态错误: current=%d, limit=%d", current, limit)
	}
}

// TestFixedWindowCounter_MultipleWindows 测试多个窗口周期
func TestFixedWindowCounter_MultipleWindows(t *testing.T) {
	limiter := NewFixedWindowCounter(2, 50*time.Millisecond)
	
	// 第一个窗口
	limiter.Allow()
	limiter.Allow()
	if limiter.Allow() {
		t.Error("第一个窗口第3个请求应该被拒绝")
	}
	
	// 等待进入第二个窗口
	time.Sleep(60 * time.Millisecond)
	
	// 第二个窗口
	if !limiter.Allow() {
		t.Error("第二个窗口第1个请求应该通过")
	}
	if !limiter.Allow() {
		t.Error("第二个窗口第2个请求应该通过")
	}
	if limiter.Allow() {
		t.Error("第二个窗口第3个请求应该被拒绝")
	}
}

// BenchmarkFixedWindowCounter_Allow 性能测试
func BenchmarkFixedWindowCounter_Allow(b *testing.B) {
	limiter := NewFixedWindowCounter(int64(b.N), time.Hour)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			limiter.Allow()
		}
	})
}

// BenchmarkFixedWindowCounter_GetStatus 状态获取性能测试
func BenchmarkFixedWindowCounter_GetStatus(b *testing.B) {
	limiter := NewFixedWindowCounter(1000, time.Hour)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			limiter.GetStatus()
		}
	})
}