package limit

import (
	"sync"
	"testing"
	"time"
)

// TestLeakyBucket_Basic 测试漏桶基本功能
func TestLeakyBucket_Basic(t *testing.T) {
	bucket := NewLeakyBucket(3, 100*time.Millisecond) // 容量3，每100ms漏一个
	defer bucket.Stop()
	
	// 应该可以添加3个请求
	for i := 0; i < 3; i++ {
		if !bucket.Allow() {
			t.Errorf("第%d个请求应该可以添加", i+1)
		}
	}
	
	// 第4个请求应该失败（桶满了）
	if bucket.Allow() {
		t.Error("第4个请求应该失败")
	}
	
	// 检查状态
	current, capacity := bucket.GetStatus()
	if current != 3 || capacity != 3 {
		t.Errorf("状态错误: current=%d, capacity=%d", current, capacity)
	}
}

// TestLeakyBucket_Leak 测试漏桶功能
func TestLeakyBucket_Leak(t *testing.T) {
	bucket := NewLeakyBucket(2, 50*time.Millisecond) // 容量2，每50ms漏一个
	defer bucket.Stop()
	
	// 填满桶
	bucket.Allow()
	bucket.Allow()
	
	// 应该满了
	if bucket.Allow() {
		t.Error("桶应该满了")
	}
	
	// 等待漏水
	time.Sleep(80 * time.Millisecond)
	
	// 现在应该可以添加请求了
	if !bucket.Allow() {
		t.Error("漏水后应该可以添加请求")
	}
	
	// 检查状态，应该有2个请求（1个漏掉了，1个新加的）
	current, _ := bucket.GetStatus()
	if current != 2 {
		t.Errorf("漏水后应该有2个请求，实际%d个", current)
	}
}

// TestLeakyBucket_AllowN 测试批量添加请求
func TestLeakyBucket_AllowN(t *testing.T) {
	bucket := NewLeakyBucket(5, 100*time.Millisecond)
	defer bucket.Stop()
	
	// 尝试添加3个请求
	if !bucket.AllowN(3) {
		t.Error("应该可以添加3个请求")
	}
	
	// 尝试添加3个请求（应该失败，只剩2个空间）
	if bucket.AllowN(3) {
		t.Error("不应该能添加3个请求")
	}
	
	// 尝试添加2个请求
	if !bucket.AllowN(2) {
		t.Error("应该可以添加2个请求")
	}
	
	// 现在桶应该满了
	if bucket.AllowN(1) {
		t.Error("桶应该满了")
	}
}

// TestLeakyBucket_ZeroRequests 测试零请求添加
func TestLeakyBucket_ZeroRequests(t *testing.T) {
	bucket := NewLeakyBucket(3, 100*time.Millisecond)
	defer bucket.Stop()
	
	// 添加0个请求应该总是成功
	if !bucket.AllowN(0) {
		t.Error("添加0个请求应该成功")
	}
	
	// 填满桶
	bucket.AllowN(3)
	
	// 即使桶满了，添加0个请求也应该成功
	if !bucket.AllowN(0) {
		t.Error("添加0个请求应该总是成功")
	}
}

// TestLeakyBucket_FastLeak 测试快速漏水
func TestLeakyBucket_FastLeak(t *testing.T) {
	bucket := NewLeakyBucket(10, 10*time.Millisecond) // 容量10，每10ms漏一个
	defer bucket.Stop()
	
	// 填满桶
	bucket.AllowN(10)
	
	// 等待一段时间
	time.Sleep(150 * time.Millisecond)
	
	// 应该漏掉了很多请求
	current, _ := bucket.GetStatus()
	if current >= 10 {
		t.Errorf("快速漏水后应该漏掉一些请求，当前还有%d个", current)
	}
}

// TestLeakyBucket_Stop 测试停止功能
func TestLeakyBucket_Stop(t *testing.T) {
	bucket := NewLeakyBucket(3, 100*time.Millisecond)
	
	// 添加一些请求
	bucket.Allow()
	bucket.Allow()
	
	// 停止桶
	bucket.Stop()
	
	// 再次停止应该不会出错
	bucket.Stop()
	
	// 停止后仍然可以添加请求（但不会漏水）
	if !bucket.Allow() {
		t.Error("停止后应该仍可以添加请求")
	}
}

// TestLeakyBucket_Concurrent 测试并发安全性
func TestLeakyBucket_Concurrent(t *testing.T) {
	bucket := NewLeakyBucket(50, 1*time.Millisecond) // 容量50，快速漏水
	defer bucket.Stop()
	
	var wg sync.WaitGroup
	var successCount int64
	var mu sync.Mutex
	
	// 启动100个并发请求
	for i := 0; i < 100; i++ {
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
	
	// 由于快速漏水，应该有一定数量的请求成功
	if successCount == 0 {
		t.Error("并发测试中应该有一些请求成功")
	}
	
	if successCount > 50 {
		t.Errorf("成功请求数不应该超过容量: %d", successCount)
	}
}

// TestLeakyBucket_ConcurrentAllowN 测试并发AllowN
func TestLeakyBucket_ConcurrentAllowN(t *testing.T) {
	bucket := NewLeakyBucket(100, 10*time.Millisecond)
	defer bucket.Stop()
	
	var wg sync.WaitGroup
	var totalAdded int64
	var mu sync.Mutex
	
	// 启动10个并发请求，每个尝试添加10个请求
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if bucket.AllowN(10) {
				mu.Lock()
				totalAdded += 10
				mu.Unlock()
			}
		}()
	}
	
	wg.Wait()
	
	// 总添加数不应该超过容量
	if totalAdded > 100 {
		t.Errorf("总添加数超过容量: %d", totalAdded)
	}
}

// TestLeakyBucket_EmptyBucketLeak 测试空桶漏水
func TestLeakyBucket_EmptyBucketLeak(t *testing.T) {
	bucket := NewLeakyBucket(3, 50*time.Millisecond)
	defer bucket.Stop()
	
	// 不添加任何请求，等待一段时间
	time.Sleep(200 * time.Millisecond)
	
	// 桶应该仍然是空的
	current, _ := bucket.GetStatus()
	if current != 0 {
		t.Errorf("空桶漏水后应该仍然是空的，实际%d个", current)
	}
	
	// 应该可以添加请求
	if !bucket.Allow() {
		t.Error("空桶应该可以添加请求")
	}
}

// TestLeakyBucket_LeakToEmpty 测试漏到空桶
func TestLeakyBucket_LeakToEmpty(t *testing.T) {
	bucket := NewLeakyBucket(2, 30*time.Millisecond)
	defer bucket.Stop()
	
	// 添加2个请求
	bucket.AllowN(2)
	
	// 等待足够时间让所有请求漏完
	time.Sleep(100 * time.Millisecond)
	
	// 桶应该是空的
	current, _ := bucket.GetStatus()
	if current != 0 {
		t.Errorf("所有请求漏完后桶应该是空的，实际%d个", current)
	}
	
	// 应该可以添加满桶的请求
	if !bucket.AllowN(2) {
		t.Error("空桶应该可以添加满桶的请求")
	}
}

// BenchmarkLeakyBucket_Allow 性能测试
func BenchmarkLeakyBucket_Allow(b *testing.B) {
	bucket := NewLeakyBucket(int64(b.N), time.Nanosecond) // 极快的漏水速率
	defer bucket.Stop()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bucket.Allow()
		}
	})
}

// BenchmarkLeakyBucket_AllowN 批量添加性能测试
func BenchmarkLeakyBucket_AllowN(b *testing.B) {
	bucket := NewLeakyBucket(int64(b.N*10), time.Nanosecond)
	defer bucket.Stop()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bucket.AllowN(10)
		}
	})
}

// BenchmarkLeakyBucket_GetStatus 状态获取性能测试
func BenchmarkLeakyBucket_GetStatus(b *testing.B) {
	bucket := NewLeakyBucket(1000, 100*time.Millisecond)
	defer bucket.Stop()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bucket.GetStatus()
		}
	})
}