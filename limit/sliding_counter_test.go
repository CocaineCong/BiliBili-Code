package limit

import (
	"sync"
	"testing"
	"time"
)

// TestSlidingWindowCounter_Basic 测试滑动窗口计数器基本功能
func TestSlidingWindowCounter_Basic(t *testing.T) {
	limiter := NewSlidingWindowCounter(3, time.Second, 100*time.Millisecond)
	
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

// TestSlidingWindowCounter_SlidingWindow 测试滑动窗口特性
func TestSlidingWindowCounter_SlidingWindow(t *testing.T) {
	limiter := NewSlidingWindowCounter(2, 200*time.Millisecond, 50*time.Millisecond)
	
	// 发送2个请求
	limiter.Allow()
	limiter.Allow()
	
	// 应该被拒绝
	if limiter.Allow() {
		t.Error("请求应该被拒绝")
	}
	
	// 等待一半窗口时间，让部分请求过期
	time.Sleep(120 * time.Millisecond)
	
	// 现在应该可以通过一些请求
	if !limiter.Allow() {
		t.Error("滑动窗口后请求应该通过")
	}
}

// TestSlidingWindowCounter_Precision 测试精度设置
func TestSlidingWindowCounter_Precision(t *testing.T) {
	// 高精度滑动窗口
	highPrecision := NewSlidingWindowCounter(5, time.Second, 10*time.Millisecond)
	
	// 低精度滑动窗口
	lowPrecision := NewSlidingWindowCounter(5, time.Second, 200*time.Millisecond)
	
	// 发送请求
	for i := 0; i < 3; i++ {
		highPrecision.Allow()
		lowPrecision.Allow()
	}
	
	// 检查状态
	highCurrent, _ := highPrecision.GetStatus()
	lowCurrent, _ := lowPrecision.GetStatus()
	
	if highCurrent != 3 || lowCurrent != 3 {
		t.Errorf("精度测试失败: high=%d, low=%d", highCurrent, lowCurrent)
	}
}

// TestSlidingWindowCounter_WindowExpiry 测试窗口过期
func TestSlidingWindowCounter_WindowExpiry(t *testing.T) {
	limiter := NewSlidingWindowCounter(3, 100*time.Millisecond, 20*time.Millisecond)
	
	// 填满限制
	for i := 0; i < 3; i++ {
		limiter.Allow()
	}
	
	// 应该被拒绝
	if limiter.Allow() {
		t.Error("请求应该被拒绝")
	}
	
	// 等待窗口完全过期
	time.Sleep(150 * time.Millisecond)
	
	// 现在应该可以通过
	if !limiter.Allow() {
		t.Error("窗口过期后请求应该通过")
	}
	
	// 检查状态，应该只有1个请求
	current, _ := limiter.GetStatus()
	if current != 1 {
		t.Errorf("窗口过期后计数应该为1，实际为%d", current)
	}
}

// TestSlidingWindowCounter_Concurrent 测试并发安全性
func TestSlidingWindowCounter_Concurrent(t *testing.T) {
	limiter := NewSlidingWindowCounter(100, time.Second, 10*time.Millisecond)
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
}

// TestSlidingWindowCounter_ZeroLimit 测试零限制
func TestSlidingWindowCounter_ZeroLimit(t *testing.T) {
	limiter := NewSlidingWindowCounter(0, time.Second, 100*time.Millisecond)
	
	// 任何请求都应该被拒绝
	if limiter.Allow() {
		t.Error("零限制时请求应该被拒绝")
	}
	
	current, limit := limiter.GetStatus()
	if current != 0 || limit != 0 {
		t.Errorf("零限制状态错误: current=%d, limit=%d", current, limit)
	}
}

// TestSlidingWindowCounter_CleanExpiredWindows 测试过期窗口清理
func TestSlidingWindowCounter_CleanExpiredWindows(t *testing.T) {
	limiter := NewSlidingWindowCounter(10, 100*time.Millisecond, 20*time.Millisecond)
	
	// 发送一些请求
	for i := 0; i < 5; i++ {
		limiter.Allow()
	}
	
	// 等待窗口过期
	time.Sleep(150 * time.Millisecond)
	
	// 发送新请求，这应该触发清理
	limiter.Allow()
	
	// 检查状态，应该只有1个请求（新的）
	current, _ := limiter.GetStatus()
	if current != 1 {
		t.Errorf("清理后计数应该为1，实际为%d", current)
	}
}

// BenchmarkSlidingWindowCounter_Allow 性能测试
func BenchmarkSlidingWindowCounter_Allow(b *testing.B) {
	limiter := NewSlidingWindowCounter(int64(b.N), time.Hour, time.Second)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			limiter.Allow()
		}
	})
}

// BenchmarkSlidingWindowCounter_GetStatus 状态获取性能测试
func BenchmarkSlidingWindowCounter_GetStatus(b *testing.B) {
	limiter := NewSlidingWindowCounter(1000, time.Hour, time.Second)
	
	// 预先添加一些请求
	for i := 0; i < 100; i++ {
		limiter.Allow()
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			limiter.GetStatus()
		}
	})
}