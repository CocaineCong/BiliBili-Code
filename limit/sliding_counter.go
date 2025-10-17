package limit

import (
	"sync"
	"time"
)

// SlidingWindowCounter 滑动窗口计数器限流器
type SlidingWindowCounter struct {
	limit     int64           // 限制数量
	window    time.Duration   // 时间窗口
	requests  map[int64]int64 // 时间戳到请求数的映射
	mutex     sync.Mutex      // 互斥锁
	precision time.Duration   // 精度（子窗口大小）
}

// NewSlidingWindowCounter 创建滑动窗口计数器
func NewSlidingWindowCounter(limit int64, window time.Duration, precision time.Duration) *SlidingWindowCounter {
	return &SlidingWindowCounter{
		limit:     limit,
		window:    window,
		requests:  make(map[int64]int64),
		precision: precision,
	}
}

// Allow 检查是否允许请求通过
func (s *SlidingWindowCounter) Allow() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	now := time.Now()
	currentWindow := now.Truncate(s.precision).Unix()
	s.cleanExpiredWindows(now)                    // 清理过期的窗口数据
	totalRequests := s.countRequestsInWindow(now) // 计算当前窗口内的总请求数
	if totalRequests >= s.limit {                 // 检查是否超过限制
		return false
	}
	s.requests[currentWindow]++ // 增加当前窗口的计数
	return true
}

// cleanExpiredWindows 清理过期的窗口数据
func (s *SlidingWindowCounter) cleanExpiredWindows(now time.Time) {
	cutoff := now.Add(-s.window).Unix()
	for timestamp := range s.requests {
		if timestamp < cutoff {
			delete(s.requests, timestamp)
		}
	}
}

// countRequestsInWindow 计算窗口内的请求总数
func (s *SlidingWindowCounter) countRequestsInWindow(now time.Time) int64 {
	cutoff := now.Add(-s.window)
	total := int64(0)

	for timestamp, count := range s.requests {
		windowTime := time.Unix(timestamp, 0)
		if windowTime.After(cutoff) {
			total += count
		}
	}

	return total
}

// GetStatus 获取当前状态
func (s *SlidingWindowCounter) GetStatus() (int64, int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	s.cleanExpiredWindows(now)
	current := s.countRequestsInWindow(now)

	return current, s.limit
}
