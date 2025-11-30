package limit

import (
	"fmt"
	"testing"
	"time"
)

// TestRateLimiterInterface 测试所有限流器都实现了RateLimiter接口
func TestRateLimiterInterface(t *testing.T) {
	testCases := []struct {
		name    string
		limiter RateLimiter
		cleanup func()
	}{
		{
			name:    "FixedWindowCounter",
			limiter: NewFixedWindowCounter(5, time.Second),
			cleanup: func() {},
		},
		{
			name:    "SlidingWindowCounter",
			limiter: NewSlidingWindowCounter(5, time.Second, 100*time.Millisecond),
			cleanup: func() {},
		},
		{
			name:    "TokenBucket",
			limiter: NewTokenBucket(5, 1),
			cleanup: func() {
				if tb, ok := interface{}(NewTokenBucket(5, 1)).(*TokenBucket); ok {
					tb.Stop()
				}
			},
		},
		{
			name:    "LeakyBucket",
			limiter: NewLeakyBucket(5, 100*time.Millisecond),
			cleanup: func() {
				if lb, ok := interface{}(NewLeakyBucket(5, 100*time.Millisecond)).(*LeakyBucket); ok {
					lb.Stop()
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer tc.cleanup()

			// 测试Allow方法
			if !tc.limiter.Allow() {
				t.Errorf("%s: Allow方法应该返回true", tc.name)
			}

			// 测试GetStatus方法
			current, limit := tc.limiter.GetStatus()
			if current < 0 || limit <= 0 {
				t.Errorf("%s: GetStatus返回值异常: current=%d, limit=%d", tc.name, current, limit)
			}

			// 测试接口类型断言
			var rateLimiter RateLimiter = tc.limiter
			if rateLimiter == nil {
				t.Errorf("%s: 应该实现RateLimiter接口", tc.name)
			}
		})
	}
}

// TestRateLimiterBehavior 测试所有限流器的基本行为一致性

// TestRateLimiterZeroLimit 测试零限制的一致性
func TestRateLimiterZeroLimit(t *testing.T) {
	limiters := map[string]RateLimiter{
		"FixedWindowCounter":   NewFixedWindowCounter(0, time.Second),
		"SlidingWindowCounter": NewSlidingWindowCounter(0, time.Second, 100*time.Millisecond),
		"TokenBucket":          NewTokenBucket(0, 1),
		"LeakyBucket":          NewLeakyBucket(0, 100*time.Millisecond),
	}

	// 清理资源
	defer func() {
		if tb, ok := limiters["TokenBucket"].(*TokenBucket); ok {
			tb.Stop()
		}
		if lb, ok := limiters["LeakyBucket"].(*LeakyBucket); ok {
			lb.Stop()
		}
	}()

	for name, limiter := range limiters {
		t.Run(fmt.Sprintf("%s_零限制", name), func(t *testing.T) {
			// 任何请求都应该被拒绝
			if limiter.Allow() {
				t.Errorf("%s: 零限制时请求应该被拒绝", name)
			}

			// 状态检查
			current, limit := limiter.GetStatus()
			if current != 0 || limit != 0 {
				t.Errorf("%s: 零限制状态错误: current=%d, limit=%d", name, current, limit)
			}
		})
	}
}

// BenchmarkRateLimiters 比较所有限流器的性能
func BenchmarkRateLimiters(b *testing.B) {
	benchmarks := []struct {
		name    string
		limiter RateLimiter
		cleanup func()
	}{
		{
			name:    "FixedWindowCounter",
			limiter: NewFixedWindowCounter(int64(b.N), time.Hour),
			cleanup: func() {},
		},
		{
			name:    "SlidingWindowCounter",
			limiter: NewSlidingWindowCounter(int64(b.N), time.Hour, time.Second),
			cleanup: func() {},
		},
		{
			name:    "TokenBucket",
			limiter: NewTokenBucket(int64(b.N), 1000000),
			cleanup: func() {
				if tb, ok := interface{}(NewTokenBucket(int64(b.N), 1000000)).(*TokenBucket); ok {
					tb.Stop()
				}
			},
		},
		{
			name:    "LeakyBucket",
			limiter: NewLeakyBucket(int64(b.N), time.Nanosecond),
			cleanup: func() {
				if lb, ok := interface{}(NewLeakyBucket(int64(b.N), time.Nanosecond)).(*LeakyBucket); ok {
					lb.Stop()
				}
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			defer bm.cleanup()

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					bm.limiter.Allow()
				}
			})
		})
	}
}
