package limit

import (
	"sync"
	"testing"
	"time"
)

// TestLeakyBucket_Basic 测试漏桶基本功能（Traffic Shaping 模式）
func TestLeakyBucket_Basic(t *testing.T) {
	// 容量3，每100ms漏一个
	bucket := NewLeakyBucket(3, 100*time.Millisecond)
	defer bucket.Stop()
	// 串行调用 Allow()
	// 第1个：立即返回
	start := time.Now()
	if !bucket.Allow() {
		t.Error("第1个请求应该成功")
	}
	// 第2个：应该阻塞约 100ms
	if !bucket.Allow() {
		t.Error("第2个请求应该成功")
	}
	// 第3个：应该再阻塞约 100ms
	if !bucket.Allow() {
		t.Error("第3个请求应该成功")
	}
	elapsed := time.Since(start)
	// 总耗时应该至少 200ms (第2个等待100ms，第3个等待100ms)
	if elapsed < 180*time.Millisecond {
		t.Errorf("流量整形未生效，3个请求总耗时过短: %v", elapsed)
	}
}

// TestLeakyBucket_Overflow 测试桶满拒绝
func TestLeakyBucket_Overflow(t *testing.T) {
	// 容量2，每100ms漏一个
	bucket := NewLeakyBucket(2, 100*time.Millisecond)
	defer bucket.Stop()

	// 模拟并发突发流量
	var wg sync.WaitGroup
	results := make(chan bool, 5)

	// 同时发起 5 个请求
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- bucket.Allow()
		}()
	}

	wg.Wait()
	close(results)

	successCount := 0
	for res := range results {
		if res {
			successCount++
		}
	}

	// 注意：由于并发调度的不确定性，可能有些请求稍微晚一点进入，导致 now 增加，从而可能通过更多。
	// 但在极短时间内，应该接近 2 个。
	if successCount < 2 || successCount > 3 {
		t.Errorf("预期通过2-3个请求，实际通过 %d 个", successCount)
	}
}

// TestLeakyBucket_AllowN 测试批量请求
func TestLeakyBucket_AllowN(t *testing.T) {
	bucket := NewLeakyBucket(10, 50*time.Millisecond)
	defer bucket.Stop()

	start := time.Now()

	// 请求 3 个，相当于 3 * 50ms = 150ms 的水量
	// 第1次：立即成功，但水位增加 150ms
	if !bucket.AllowN(3) {
		t.Error("AllowN(3) 应该成功")
	}

	// 紧接着再请求 1 个
	// 应该阻塞 150ms (等待前面的 3 个漏完)
	if !bucket.AllowN(1) {
		t.Error("AllowN(1) 应该成功")
	}

	elapsed := time.Since(start)
	if elapsed < 140*time.Millisecond {
		t.Errorf("AllowN 应该导致后续请求阻塞，实际耗时: %v", elapsed)
	}
}

// TestLeakyBucket_GetStatus 测试状态获取
func TestLeakyBucket_GetStatus(t *testing.T) {
	bucket := NewLeakyBucket(5, 100*time.Millisecond)
	defer bucket.Stop()

	// 初始状态
	cur, cap := bucket.GetStatus()
	if cur != 0 || cap != 5 {
		t.Errorf("初始状态错误: %d/%d", cur, cap)
	}
	// 添加一个请求
	bucket.Allow()
	// 立即检查状态，应该有 1 个在排队
	cur, _ = bucket.GetStatus()
	if cur != 1 {
		t.Errorf("添加1个请求后，状态应该是1，实际 %d", cur)
	}

	// 等待漏水
	time.Sleep(110 * time.Millisecond)
	cur, _ = bucket.GetStatus()
	if cur != 0 {
		t.Errorf("漏水后状态应该是0，实际 %d", cur)
	}
}

// TestLeakyBucket_TrafficShaping 验证流量整形效果
func TestLeakyBucket_TrafficShaping(t *testing.T) {
	// 速率：每 50ms 一个
	bucket := NewLeakyBucket(10, 50*time.Millisecond)
	defer bucket.Stop()

	start := time.Now()
	count := 5

	// 连续发送 5 个请求
	for i := 0; i < count; i++ {
		bucket.Allow()
	}

	elapsed := time.Since(start)
	// 5 个请求，第一个立即，后 4 个各等待 50ms
	// 总耗时约 4 * 50 = 200ms
	expected := 200 * time.Millisecond

	if elapsed < expected-20*time.Millisecond {
		t.Errorf("流量整形过快: %v, 预期 >= %v", elapsed, expected)
	}
}

// BenchmarkLeakyBucket_Allow 性能测试
func BenchmarkLeakyBucket_Allow(b *testing.B) {
	bucket := NewLeakyBucket(int64(b.N), time.Nanosecond) // 极快的漏水速率
	defer bucket.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bucket.Allow()
	}
}
