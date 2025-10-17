package limit

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow() bool
	GetStatus() (int64, int64)
}
