package limit

import (
	"sync"
	"time"
)

// LeakyBucket 漏桶算法实现
type LeakyBucket struct {
	capacity  int64         // 桶容量
	tokens    int64         // 当前桶中的令牌数
	leakRate  time.Duration // 漏桶速率（多久漏一个令牌）
	lastLeak  time.Time     // 上次漏桶时间
	mutex     sync.Mutex    // 互斥锁
	stopCh    chan struct{} // 停止信号
	isRunning bool          // 是否正在运行
}

// NewLeakyBucket 创建新的漏桶
// capacity: 桶容量
// leakRate: 漏桶速率，例如 100*time.Millisecond 表示每100毫秒漏一个令牌
func NewLeakyBucket(capacity int64, leakRate time.Duration) *LeakyBucket {
	bucket := &LeakyBucket{
		capacity:  capacity,
		tokens:    0,
		leakRate:  leakRate,
		lastLeak:  time.Now(),
		stopCh:    make(chan struct{}),
		isRunning: false,
	}

	// 启动漏桶协程
	bucket.start()
	return bucket
}

// start 启动漏桶的定时漏水协程
func (lb *LeakyBucket) start() {
	lb.mutex.Lock()
	if lb.isRunning {
		lb.mutex.Unlock()
		return
	}
	lb.isRunning = true
	lb.mutex.Unlock()

	go func() {
		ticker := time.NewTicker(lb.leakRate)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				lb.leak()
			case <-lb.stopCh:
				return
			}
		}
	}()
}

// leak 执行漏桶操作
func (lb *LeakyBucket) leak() {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	if lb.tokens > 0 {
		lb.tokens--
		lb.lastLeak = time.Now()
	}
}

// Allow 尝试向桶中添加一个请求
// 如果桶未满，返回 true；如果桶满了，返回 false
func (lb *LeakyBucket) Allow() bool {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	if lb.tokens < lb.capacity {
		lb.tokens++
		return true
	}
	return false
}

// AllowN 尝试向桶中添加 n 个请求
func (lb *LeakyBucket) AllowN(n int64) bool {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	if lb.tokens+n <= lb.capacity {
		lb.tokens += n
		return true
	}
	return false
}

// GetStatus 获取当前桶的状态
func (lb *LeakyBucket) GetStatus() (current int64, capacity int64) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	return lb.tokens, lb.capacity
}

// Stop 停止漏桶
func (lb *LeakyBucket) Stop() {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	if lb.isRunning {
		close(lb.stopCh)
		lb.isRunning = false
	}
}
