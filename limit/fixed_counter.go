package limit

import (
	"sync"
	"time"
)

// FixedWindowCounter 固定窗口计数器限流器
type FixedWindowCounter struct {
	limit    int64         // 限制数量
	window   time.Duration // 时间窗口
	counter  int64         // 当前计数
	lastTime time.Time     // 上次重置时间
	mutex    sync.Mutex    // 互斥锁
}

// NewFixedWindowCounter 创建固定窗口计数器
func NewFixedWindowCounter(limit int64, window time.Duration) *FixedWindowCounter {
	return &FixedWindowCounter{
		limit:    limit,
		window:   window,
		counter:  0,
		lastTime: time.Now(),
	}
}

// Allow 检查是否允许请求通过
func (f *FixedWindowCounter) Allow() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	now := time.Now()
	// 如果超过了时间窗口，重置计数器
	if now.Sub(f.lastTime) >= f.window {
		f.counter = 0
		f.lastTime = now
	}
	// 检查是否超过限制
	if f.counter >= f.limit {
		return false
	}
	f.counter++
	return true
}

// GetStatus 获取当前状态
func (f *FixedWindowCounter) GetStatus() (int64, int64) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.counter, f.limit
}
