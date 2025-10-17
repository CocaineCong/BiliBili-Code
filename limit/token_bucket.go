package limit

import (
	"sync"
	"time"
)

// TokenBucket 令牌桶算法实现
type TokenBucket struct {
	capacity     int64         // 桶容量（最大令牌数）
	tokens       int64         // 当前令牌数
	refillRate   int64         // 令牌补充速率（每秒补充多少个令牌）
	refillPeriod time.Duration // 补充周期
	lastRefill   time.Time     // 上次补充时间
	mutex        sync.Mutex    // 互斥锁
	stopCh       chan struct{} // 停止信号
	isRunning    bool          // 是否正在运行
}

// NewTokenBucket 创建新的令牌桶
// capacity: 桶容量
// refillRate: 每秒补充的令牌数
func NewTokenBucket(capacity int64, refillRate int64) *TokenBucket {
	bucket := &TokenBucket{
		capacity:     capacity,
		tokens:       capacity, // 初始时桶是满的
		refillRate:   refillRate,
		refillPeriod: time.Second / time.Duration(refillRate), // 计算每个令牌的补充间隔
		lastRefill:   time.Now(),
		stopCh:       make(chan struct{}),
		isRunning:    false,
	}

	// 启动令牌补充协程
	bucket.start()
	return bucket
}

// start 启动令牌补充协程
func (tb *TokenBucket) start() {
	tb.mutex.Lock()
	if tb.isRunning {
		tb.mutex.Unlock()
		return
	}
	tb.isRunning = true
	tb.mutex.Unlock()

	go func() {
		ticker := time.NewTicker(tb.refillPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				tb.refill()
			case <-tb.stopCh:
				return
			}
		}
	}()
}

// refill 补充令牌
func (tb *TokenBucket) refill() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	if tb.tokens < tb.capacity {
		tb.tokens++
		tb.lastRefill = time.Now()
	}
}

// Allow 尝试获取一个令牌
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN 尝试获取 n 个令牌
func (tb *TokenBucket) AllowN(n int64) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}
	return false
}

// GetStatus 获取当前桶的状态
func (tb *TokenBucket) GetStatus() (current int64, capacity int64) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	return tb.tokens, tb.capacity
}

// Stop 停止令牌桶
func (tb *TokenBucket) Stop() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	if tb.isRunning {
		close(tb.stopCh)
		tb.isRunning = false
	}
}
