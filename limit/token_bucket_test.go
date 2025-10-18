package limit

import (
	"sync"
	"testing"
	"time"
)

// TestTokenBucket_Basic 测试令牌桶基本功能
func TestTokenBucket_Basic(t *testing.T) {
	bucket := NewTokenBucket(3, 1) // 容量3，每秒补充1个
	defer bucket.Stop()

	// 初始时桶是满的，应该可以获取3个令牌
	for i := 0; i < 3; i++ {
		if !bucket.Allow() {
			t.Errorf("第%d个令牌应该可以获取", i+1)
		}
	}

	// 第4个令牌应该获取失败
	if bucket.Allow() {
		t.Error("第4个令牌应该获取失败")
	}

	// 检查状态
	current, capacity := bucket.GetStatus()
	if current != 0 || capacity != 3 {
		t.Errorf("状态错误: current=%d, capacity=%d", current, capacity)
	}
}

// TestTokenBucket_Refill 测试令牌补充功能
func TestTokenBucket_Refill(t *testing.T) {
	bucket := NewTokenBucket(2, 10) // 容量2，每秒补充10个（每100ms一个）
	defer bucket.Stop()

	// 消耗所有令牌
	bucket.Allow()
	bucket.Allow()

	// 应该没有令牌了
	if bucket.Allow() {
		t.Error("应该没有令牌了")
	}

	// 等待令牌补充
	time.Sleep(150 * time.Millisecond)

	// 现在应该有令牌了
	if !bucket.Allow() {
		t.Error("应该有新的令牌了")
	}
}

// TestTokenBucket_AllowN 测试批量获取令牌
func TestTokenBucket_AllowN(t *testing.T) {
	bucket := NewTokenBucket(5, 1)
	defer bucket.Stop()

	// 尝试获取3个令牌
	if !bucket.AllowN(3) {
		t.Error("应该可以获取3个令牌")
	}

	// 尝试获取3个令牌（应该失败，只剩2个）
	if bucket.AllowN(3) {
		t.Error("不应该能获取3个令牌")
	}

	// 尝试获取2个令牌
	if !bucket.AllowN(2) {
		t.Error("应该可以获取2个令牌")
	}

	// 现在应该没有令牌了
	if bucket.AllowN(1) {
		t.Error("不应该还有令牌")
	}
}

// TestTokenBucket_ZeroTokens 测试零令牌获取
func TestTokenBucket_ZeroTokens(t *testing.T) {
	bucket := NewTokenBucket(5, 1)
	defer bucket.Stop()

	// 获取0个令牌应该总是成功
	if !bucket.AllowN(0) {
		t.Error("获取0个令牌应该成功")
	}

	// 消耗所有令牌
	bucket.AllowN(5)

	// 即使没有令牌，获取0个也应该成功
	if !bucket.AllowN(0) {
		t.Error("获取0个令牌应该总是成功")
	}
}

// TestTokenBucket_HighRefillRate 测试高补充速率
func TestTokenBucket_HighRefillRate(t *testing.T) {
	bucket := NewTokenBucket(10, 1000) // 容量10，每秒补充1000个
	defer bucket.Stop()

	// 消耗所有令牌
	bucket.AllowN(10)

	// 等待短时间
	time.Sleep(50 * time.Millisecond)

	// 应该有新令牌了
	if !bucket.Allow() {
		t.Error("高补充速率下应该快速补充令牌")
	}
}

// TestTokenBucket_Stop 测试停止功能
func TestTokenBucket_Stop(t *testing.T) {
	bucket := NewTokenBucket(3, 1)

	// 停止桶
	bucket.Stop()

	// 再次停止应该不会出错
	bucket.Stop()

	// 停止后仍然可以使用现有令牌
	if !bucket.Allow() {
		t.Error("停止后应该仍可以使用现有令牌")
	}
}

// TestTokenBucket_Concurrent 测试并发安全性
func TestTokenBucket_Concurrent(t *testing.T) {
	bucket := NewTokenBucket(100, 1000) // 高补充速率避免限制
	defer bucket.Stop()

	var wg sync.WaitGroup
	var successCount int64
	var mu sync.Mutex

	// 启动200个并发请求
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if bucket.Allow() {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// 由于高补充速率，大部分请求应该成功
	if successCount < 100 {
		t.Errorf("并发测试中成功请求太少: %d", successCount)
	}
}

// TestTokenBucket_ConcurrentAllowN 测试并发AllowN
func TestTokenBucket_ConcurrentAllowN(t *testing.T) {
	bucket := NewTokenBucket(100, 1)
	defer bucket.Stop()

	var wg sync.WaitGroup
	var totalConsumed int64
	var mu sync.Mutex

	// 启动10个并发请求，每个尝试获取10个令牌
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if bucket.AllowN(10) {
				mu.Lock()
				totalConsumed += 10
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// 总消耗不应该超过容量
	if totalConsumed > 100 {
		t.Errorf("总消耗超过容量: %d", totalConsumed)
	}
}

// TestTokenBucket_RefillLimit 测试补充上限
func TestTokenBucket_RefillLimit(t *testing.T) {
	bucket := NewTokenBucket(3, 100) // 容量3，高补充速率
	defer bucket.Stop()

	// 等待足够时间让令牌补充
	time.Sleep(100 * time.Millisecond)

	// 检查状态，令牌数不应该超过容量
	current, capacity := bucket.GetStatus()
	if current > capacity {
		t.Errorf("令牌数超过容量: current=%d, capacity=%d", current, capacity)
	}

	// 应该最多只能获取容量数量的令牌
	if !bucket.AllowN(3) {
		t.Error("应该可以获取3个令牌")
	}

	if bucket.AllowN(1) {
		t.Error("不应该还有令牌")
	}
}

// BenchmarkTokenBucket_Allow 性能测试
func BenchmarkTokenBucket_Allow(b *testing.B) {
	bucket := NewTokenBucket(int64(b.N), 1000000) // 高补充速率避免限制
	defer bucket.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bucket.Allow()
		}
	})
}

// BenchmarkTokenBucket_AllowN 批量获取性能测试
func BenchmarkTokenBucket_AllowN(b *testing.B) {
	bucket := NewTokenBucket(int64(b.N*10), 1000000)
	defer bucket.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bucket.AllowN(10)
		}
	})
}

// BenchmarkTokenBucket_GetStatus 状态获取性能测试
func BenchmarkTokenBucket_GetStatus(b *testing.B) {
	bucket := NewTokenBucket(1000, 1000)
	defer bucket.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bucket.GetStatus()
		}
	})
}
