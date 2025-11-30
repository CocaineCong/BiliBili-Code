package limit

import (
	"sync"
	"time"
)

// LeakyBucket 漏桶算法实现
// 1. 恒定速率流出请求
// 2. 如果桶未满但有水，请求会阻塞等待直到轮到自己
// 3. 如果桶满了，请求被拒绝
type LeakyBucket struct {
	capacity int64         // 桶容量（最大允许排队请求数）
	rate     time.Duration // 漏水速率（每个请求的处理间隔）
	lastTime time.Time     // 上一次请求的理论结束时间（水位线）
	mutex    sync.Mutex    // 互斥锁
}

// NewLeakyBucket 创建新的漏桶
// capacity: 桶容量
// leakRate: 漏桶速率，例如 100 表示每100毫秒漏一个令牌
func NewLeakyBucket(capacity int64, leakRate time.Duration) *LeakyBucket {
	return &LeakyBucket{
		capacity: capacity,
		rate:     leakRate,
		lastTime: time.Now(),
	}
}

// Allow 尝试向桶中添加一个请求
// 如果桶未满，该方法会阻塞直到请求被漏出
// 如果桶满了，立即返回 false
func (lb *LeakyBucket) Allow() bool {
	return lb.AllowN(1)
}

// AllowN 尝试向桶中添加 n 个请求
func (lb *LeakyBucket) AllowN(n int64) bool {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	now := time.Now()
	if now.After(lb.lastTime) {
		lb.lastTime = now
	}
	increment := lb.rate * time.Duration(n)
	newLastTime := lb.lastTime.Add(increment)
	// 检查是否超过容量
	maxWait := lb.rate * time.Duration(lb.capacity)
	if newLastTime.Sub(now) > maxWait {
		return false
	}
	// 计算当前请求需要等待的时间（排在前面的请求处理完的时间）
	waitTime := lb.lastTime.Sub(now)
	lb.lastTime = newLastTime
	// 如果需要等待，则阻塞
	if waitTime > 0 {
		time.Sleep(waitTime)
	}

	return true
}

// GetStatus 获取当前桶的状态
// current: 当前桶中的排队请求数
// capacity: 桶的总容量
func (lb *LeakyBucket) GetStatus() (current int64, capacity int64) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	now := time.Now()
	if now.After(lb.lastTime) {
		return 0, lb.capacity
	}
	// 计算剩余排队时间
	remainingTime := lb.lastTime.Sub(now)
	// 转换为请求数，向上取整
	// 如果还有剩余时间，说明桶里有水
	if remainingTime > 0 {
		current = int64(remainingTime / lb.rate)
		if remainingTime%lb.rate != 0 {
			current++
		}
	}

	return current, lb.capacity
}

// Stop 停止漏桶
func (lb *LeakyBucket) Stop() {}
